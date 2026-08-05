# Vine Redis, RDB, Logging, and Testing

This reference is based on Vine `v0.12.0`. Use it for infrastructure Components, sensitive logging, and `app/testkit` tasks.

## Contents

- [Redis Component](#redis-component)
- [Redis locks and Cache](#redis-locks-and-cache)
- [RDB Component](#rdb-component)
- [Logging and redaction](#logging-and-redaction)
- [Testkit and validation](#testkit-and-validation)

## Redis Component

```go
type User struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type UserCache struct{ redis.Cache[*User] }
type UserLocker struct{ redis.Locker }
type MainRedis struct{ redis.Redis }

func (*MainRedis) InitOption(option *redis.Option) {
    option.Endpoint = "redis://127.0.0.1:6379/0"
}

func (*MainRedis) InitLockers(add redis.TypeAdder) {
    add(reflect.TypeFor[*UserLocker]())
}

func (*MainRedis) InitCaches(add redis.TypeAdder) {
    add(reflect.TypeFor[*UserCache]())
}

func (*DemoApp) InitComponents(add app.TypeAdder) {
    add(app.T[*MainRedis]())
}
```

Inject `*MainRedis` for ordinary go-redis commands, or inject dedicated Cache and Locker types. Vine factories capture the current execution context; do not cache these objects into a longer lifecycle.

By default, the full Go type derives the namespace. Override `KeyPrefix()` with the same value only when different types genuinely need to share one keyspace. `GetOrLoad` is not singleflight; high-concurrency misses require application-level stampede control.

## Redis Locks and Cache

A lock is a coordination lease with a TTL and refresh, not a strongly consistent ownership token, and it has no fencing token.

```go
lock, ok := service.Locker.Lock(userID)
if !ok {
    return
}

if !service.rebuildWhileOwned(lock.Context(), userID) {
    return
}

// Ownership may still be lost after the work returns. Release through an
// application-owned fail-fast or recovery boundary; do not treat IsBroken()
// as proof that Unlock() is safe.
service.releaseThroughAuditedBoundary(lock)
```

The `releaseThroughAuditedBoundary` above represents a failure policy selected and tested by the application; it is not a Vine API. It may preserve `Unlock()` panic as fail-fast behavior, or catch, log, and convert it only at an explicit recovery boundary. Do not pretend inside business logic that release always succeeds.

High-risk rules:

- `Lock.Context()` is canceled on unlock, ownership loss, refresh failure, parent cancellation, or local conservative lease expiry.
- A broken lock no longer belongs to the caller, and `Unlock()` panics.
- `IsBroken()` and the following `Unlock()` are not atomic; ownership can still be lost between them.
- The current API has no `TryUnlock`.
- Do not unconditionally `defer lock.Unlock()` when work may outlast the lease.
- Stop critical-section work immediately after context cancellation. If the business cannot accept the fail-fast unlock contract, use an explicit recovery/error boundary or choose a lock system with the required semantics.

In `v0.12.0`, the default lock lease is 30 seconds with refresh. Verify public APIs and source when the target version differs.

Common Cache operations are `Get`, `GetOrLoad`, `Set`, and `Delete`. A cache is not the business source of truth. Test keys, TTLs, misses, deserialization errors, and concurrent loads from the source.

## RDB Component

```go
type MainDatabase struct{ rdb.Database }

func (*MainDatabase) InitOption(option *rdb.Option) {
    option.ConnURL = "sqlite://./app.sqlite"
    option.MaxOpenConn = 10
}

func (*MainDatabase) InitDao(add rdb.TypeAdder) {
    add(rdb.T[*UserDao]())
}

func (*DemoApp) InitComponents(add app.TypeAdder) {
    add(app.T[*MainDatabase]())
}

type User struct {
    rdb.Model
    Name string `gorm:"column:name"`
}

type UserDao struct{ rdb.Dao[*User] }
```

Rules:

- `sqlite://...` selects SQLite. Other supported URLs use PostgreSQL.
- Components with the same `ConnURL` share the underlying `*gorm.DB`; the first opener determines pool settings.
- A DAO captures the current context and logger through `gorm.DB.WithContext`. Do not retain DAOs across executions.
- Common CRUD/query failures panic through the Vine error boundary instead of returning ordinary errors. Test the Code and status at the correct boundary.
- Use the injected DAO's exposed GORM capabilities directly for complex transactions or queries. Do not force them into a simplified API.
- `rdb.Model` includes `DeletedAt` and supports soft deletion. `rdb.DeletableModel` omits the soft-delete field and supports physical deletion.
- `standalone.Option.SQLiteFile` stores Hub state and is entirely separate from the business RDB.

Vine only opens connections and constructs DAOs. It does not call `AutoMigrate` or create business tables. Treat reviewed migrations as a separate deployment step completed before the App serves traffic. Test upgrades, rollbacks, and rolling-version compatibility.

## Logging and Redaction

Use package-level `logger.Info/Error`, or create a logger hierarchy separated by `:`:

```go
appLog := logger.New("app", "demo.user")
rpcLog := appLog.Child("rpc", "server")
rpcLog.Info("request completed", "method", "GetUser")
```

`logger` is a reserved top-level field. Level rules use literals, `*`, and `**` over complete `:`-separated segments. Rules are process-global state. Logs always go to stderr and may optionally append to a file.

Sensitive data:

- Use Skel `@sensitive` to generate a field tag or whole-type marker.
- A field named `password` is not redacted automatically.
- Dynamic maps or JSON require `redact.Option{RootSensitive: true}` or an explicit sanitizer.
- Use `RevealSensitive` only for tightly controlled diagnosis. Binary values still do not emit raw bytes.
- Framework Rpc/Event/Task payload logs always use bounded redaction and have no global disable switch.
- Error logs must not include credentials, tokens, private keys, raw configuration, unsanitized payloads, or underlying error text that may contain sensitive data.

Use `t.Cleanup` to restore global state when testing process-wide logger rules.

## Testkit and Validation

```go
func TestGreeting(t *testing.T) {
    runtime := testkit.StartStandalone[*GreetingApp](t, testkit.Option{})
    execution := runtime.NewExecution(testkit.ExecutionOption{
        Actor: meta.NewAnonymousActor(),
    })
    client := testkit.NewClient[skeled.GreetingServiceClient](execution)

    got := client.Hello("Vine")
    require.Equal(t, "Hello, Vine", got.Message)
}
```

Use configuration overrides to construct strict test data. Use the corresponding ER client when an explicit SystemError is needed.

Start only one standalone runtime per test process/package. Share the runtime across test groups and organize cases as subtests. App spec types and names are process singletons; repeated startup is not an isolation strategy. Do not run tests that share registries, loggers, App singletons, or in-process endpoints in parallel unless isolation has been proven.

Use `t.Cleanup` to restore:

- Global settings and registries.
- Environment variables.
- In-process endpoints.
- Background workers and temporary resources.

Test observable behavior: return values, persisted state, error codes, configuration overrides, timeouts, cancellation, duplicate delivery, graceful stop, and sensitive logs.

Standalone/testkit can validate wiring, DI, serialization, registration, and business routing semantics, but cannot prove:

- Heartbeat, leases, or sweepers.
- Process crashes or network partitions.
- DNS, firewalls, or real listeners.
- mTLS, certificates, or Portal external entry.
- Separate Link/NATS/Hub restart recovery.

In application repositories, prefer the repository's own scripts, then run:

```bash
gofmt -w <changed-go-files>
go test ./path/to/package
go test ./...
go vet ./...
git diff --check
```

Use the following repository-wide entry points only when changing the Vine framework repository itself:

```bash
bash test/test.sh
bash test/race.sh
VINE_RACE_SCOPE=all bash test/race.sh
GOWORK=off go vet ./...
```

Changes to public APIs, reflection, concurrency, runtime wiring, or resource lifecycle require race/vet and targeted failure tests. See [runtime-operations.md](runtime-operations.md) for networking and security semantics.
