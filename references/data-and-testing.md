# Vine Redis, RDB, Logging, and Testing

This reference is based on Vine `v0.32.5`. Use it for infrastructure Components, sensitive logging, and `app/testkit` tasks. Values such as the lock lease here are `v0.32.5` version facts; the authoritative list is in [Version baseline quick reference](foundations.md#version-baseline-quick-reference).

## Contents

- [Redis Component](#redis-component)
- [Runtime-managed locks](#runtime-managed-locks)
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

Inject `*MainRedis` for ordinary go-redis commands, or inject dedicated Cache and Locker types. Vine factories capture the current execution context; use and release these objects within that execution.

By default, the full Go type derives the namespace. Give different types the same `KeyPrefix()` when they intentionally share a keyspace. `GetOrLoad` runs a loader for each cache miss; concurrent misses require application-level stampede control.

Memory Redis (`redis+memory://name[/dbIndex]`) is process-local and ephemeral. Each process owns independent data and locks for its memory URLs; cross-process coordination requires a shared backend. External components share pools by endpoint and reject `SELECT` to a different configured database; use distinct endpoints for distinct databases.

## Runtime-Managed Locks

`core/lock.Locker` scopes keys to an application; `core/lock.UniversalLocker` shares a namespace across applications using the same runtime backend. Hub defaults to embedded leases and network deployments can select Redis or disable locking. Choose this separately from an application-owned `infra/redis.Locker`. Verify the selected backend and its lease/cancellation contract. Establish fencing and durable ownership through an explicit coordination design when required.

## Redis Locks and Cache

A lock is a coordination lease with a TTL and refresh. Strong ownership enforcement requires application-managed fencing at the protected resource.

```go
lock, ok := service.Locker.Lock(userID)
if !ok { return }
defer lock.TryUnlock()
service.rebuildWhileOwned(lock.Context(), userID)
```

`TryUnlock()` checks local state atomically and attempts a token-checked Redis release. It returns false for an lock with an unacquired, broken, released or lost-ownership state; Redis command errors still panic under the infrastructure fail-fast policy. The caller must choose an appropriate error boundary.

- `Lock.Context()` is canceled on release, ownership loss, refresh failure, parent cancellation or conservative local lease expiry. Work must observe it.
- `Unlock()` panics for broken/already-released locks. `IsBroken()` followed by `Unlock()` is a race; use `TryUnlock()` when release should tolerate lost ownership.
- Implement fencing and rollback of side effects in the business coordination design when required.

In `v0.32.5`, the default lock lease is 30 seconds with refresh. Verify public APIs and source when the target version differs.

Common Cache operations are `Get`, `GetOrLoad`, `Set`, and `Delete`. Keep authoritative business state in its owning data store and treat cached values as derived data. Test keys, TTLs, misses, deserialization errors, and concurrent loads from the source.

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
- A DAO captures the current context and logger through `gorm.DB.WithContext`. Use each DAO within its owning execution.
- Common CRUD/query failures panic through the Vine error boundary. Test the Code and status at the correct boundary.
- Use the injected DAO's exposed GORM capabilities directly for complex transactions or queries.
- `rdb.Model` includes `DeletedAt` and supports soft deletion; `rdb.DeletableModel` supports physical deletion. `UModel`/`UDeletableModel` use standard-library UUIDv7 primary keys. `Dao.Filter` provides set-based writes, `Dao.One` enforces exactly one affected row, and `Exists` checks for matching rows.
- `standalone.Option.HubDBSQLiteFile` stores Hub state and is entirely separate from the business RDB.

Vine opens connections, constructs DAOs and calls each registered DAO's `EnsureSchema` before exposing DAOs. The default hook has an empty implementation; define table creation and `AutoMigrate` calls explicitly where needed. A concrete DAO may own a repeatable migration there, or the application can use a reviewed migration step. Treat reviewed migrations as a separate deployment step completed before the App serves traffic. Test upgrades, rollbacks, and rolling-version compatibility.

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
- Mark password fields explicitly with `@sensitive`.
- Dynamic maps or JSON require `redact.Option{RootSensitive: true}` or an explicit sanitizer.
- Use `RevealSensitive` for tightly controlled diagnosis. Binary values retain their summarized representation.
- Framework Rpc/Event/Task payload logs always apply bounded redaction.
- Log sanitized error codes and diagnostic fields. Redact credentials, tokens, private keys, configuration and payloads, and sanitize underlying error text before logging.

Use `t.Cleanup` to restore global state when testing process-wide logger rules.

## Testkit and Validation

```go
func TestGreeting(t *testing.T) {
    runtime := testkit.StartStandalone[*GreetingApp](t, testkit.Option{})
    execution := runtime.NewExecution(testkit.ExecutionOption{
        Actor: meta.NewAnonymousActor(),
    })
    client := execution.NewClient[skeled.GreetingServiceClient]()

    got := client.Hello("Vine")
    require.Equal(t, "Hello, Vine", got.Message)
}
```

This testkit client path is for backend `pub service` contracts. For API services, generate an external client and call real Portal, or test the business service directly. The server example uses real Portal integration tests. Use `testkit.Option.SeedHubDataFile` or configuration overrides to construct strict test data. Use the corresponding ER client when an explicit SystemError is needed.

Start only one standalone runtime per test process/package. Share the runtime across test groups and organize cases as subtests. App spec types and names are process singletons. Run tests sharing registries, loggers, App singletons or in-process endpoints sequentially; use proven isolation for parallel tests.

Use `t.Cleanup` to restore:

- Global settings and registries.
- Environment variables.
- In-process endpoints.
- Background workers and temporary resources.

Test observable behavior: return values, persisted state, error codes, configuration overrides, timeouts, cancellation, duplicate delivery, graceful stop, and sensitive logs.

Standalone/testkit validates wiring, DI, in-process isolation, registration and business behavior. A deliberately configured real Portal listener can also prove local admission, as the example does. Use an actual network topology and dedicated checks to validate:

- Heartbeat, leases, or sweepers.
- Process crashes or network partitions.
- DNS, firewalls, or real listeners.
- mTLS, certificates and actual Portal admission.
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
