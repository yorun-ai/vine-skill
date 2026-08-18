# vine-skill Server Example Template (example-greeting)

This directory is a **verified, compilable, testable** hello-world Vine **server** example:
a single `Greeting` entity persisted in sqlite, with **actor-authentication wiring** (login
issues a session token, the Actor Auth Service validates it, and the welcome method consumes
an authenticated actor). It demonstrates the wiring steps for implementing a minimal server
with vine-skill (Go 1.26.6 / Vine v0.13.1 / skelc v0.14.0,
Rpc + actor auth + config + sqlite).

**Positioning: a pure server template.** This is a server development framework; the template
code contains **no frontend wiring** — no web capability, no Vite/proxy/serve-dist, no browser
pages. Browsers or external clients call the business Rpc through Portal RPCGW's vRPC. After
the user confirms a frontend, add the browser package under `src/web/` in the same target
project (see [frontend-vrpc.md](../../references/frontend-vrpc.md)).

When starting a new project, copy these files into the target project and replace the
placeholder names (`greeting`, `example.com/greeting`, `GreetingActor`, `GreetingConfig`, and
so on), then re-run the pinned `skelc check` / `skelc gen` gate as described in
[SKILL.md](../../SKILL.md). **Generated code (`skeled/`) is not committed; it is regenerated
by the pinned skelc.** This directory contains only hand-maintained `.skel` and `src/` files.
After copying it, generate `skeled/` with the pinned tool, initialize both
`skeled/golang/go.mod` and `src/server/go.mod`, and make the server module require the generated
module through a local `replace`. Do not treat this ungenerated directory itself as a buildable
Go module.

## Layout

```text
example-greeting/
├── README.md
├── skel/
│   ├── domain.skel              # Only the domain declaration
│   ├── actor.skel               # GreetingActor: auth { credential: token, info: username }
│   ├── auth_service.skel        # Login (noauth): login → LoginResult{token}
│   ├── welcome_service.skel     # Welcome (default auth): getWelcome → Welcome
│   ├── greeting_service.skel    # Single entity: data + service (create/list)
│   └── config.skel              # pub config GreetingConfig eternal
├── src/server/
│   ├── app/                     # App assembly + testkit tests
│   │   ├── app.go               #   Servicer registration + BindCommon + SessionStore
│   │   ├── database.go          #   RDB Component (sqlite)
│   │   ├── migrate.go           #   Migration Module: AutoMigrate in BeforeAppStart
│   │   └── greeting_test.go     #   testkit behavior tests (login + actor check + subtests)
│   ├── cmd/greeting/main.go     # standalone entry (DashboardURL decides the portal port)
│   ├── core/
│   │   ├── session.go           #   SessionStore (in-memory token→username, demo only)
│   │   ├── auth_service.go      #   login validates admin/admin123, issues token
│   │   ├── greeting_actor_auth_service.go  # generated AuthService: validates credential.token
│   │   ├── auth_service_test.go # unit test for login token → Actor Auth Service validation
│   │   ├── welcome_service.go   #   reads post-admission actor info, returns welcome
│   │   └── greeting_service.go / repository.go
│   ├── impl/                    # Generated Server adapters (embed Default...Server)
│   ├── repo/                    # DAO model + Repository implementation (skel <-> record mapping)
│   └── seed/                    # hub.yaml (config + RPCGW) + test-hub.yaml (config only)
```

After initializing the module, generating code, and resolving dependencies as described
above, `go build ./...` and `go test ./...` should both pass. This is the "minimal but
complete" pure-server reference implementation — Rpc + actor auth + config + sqlite, no
Event/Task, no frontend.

## Key Wiring Notes (Gotchas)

### 1. `.skel` file organization (`skelc check` hard rules)
- The `domain` declaration is allowed **only** in `domain.skel`, and `domain.skel` may
  contain only `@desc` plus `domain greeting`.
- Every other `.skel` file must start with a **bare `domain greeting`** (no `@desc` prefix).
- Put `pub actor` in its own file (e.g. `actor.skel`), also starting with bare `domain greeting`.
- Service ownership: `pub service XxxService { for GreetingActor via client }`. Methods that
  require login must **NOT** write `noauth` (the default is auth mode); the login method
  itself writes `noauth`.

### 2. Generation and implementation
- Generate Go for this baseline: `skelc gen go --skel-in ./skel --go-out ./skeled/golang --go-vine-version v0.13.1`. For another target project, replace the version with its reviewed pin.
- Initialize and connect the two Go modules from the project root:

```bash
(cd skeled/golang && go mod init example.com/greeting/skeled/golang)
(cd skeled/golang && go get go.yorun.ai/vine@v0.13.1 && go mod tidy)
(cd src/server && go mod init example.com/greeting/src/server)
(cd src/server && go mod edit -require=example.com/greeting/skeled/golang@v0.0.0)
(cd src/server && go mod edit -replace=example.com/greeting/skeled/golang=../../skeled/golang)
(cd src/server && go get go.yorun.ai/vine@v0.13.1 github.com/google/uuid@v1.6.0)
(cd src/server && go mod tidy)
```

- Impl Handlers **embed `Default...Server` (plain style)**; methods **return values only, no
  error**. Business errors are expressed with `ex.PanicNew(...)` in core and recovered by the
  framework into structured error responses. Use the ER variant (`Default...ServerER` + ER
  client) only when explicit error returns are needed. See `src/server/impl/greeting_service.go`
  and `src/server/core/greeting_service.go`.
- Register with `app.T[*impl.XxxService]()` via `ServicerInitHandlers`; interface-to-implementation
  bindings go in `BindCommon`; share process-wide objects with `BindInstance` (e.g. `SessionStore`).

### 3. RDB persistence (see repo/)
- Vine **does not migrate tables automatically**: add a migration Module that calls
  `GormDB().AutoMigrate(...)` in `BeforeAppStart` (see `app/migrate.go`).
- `rdb.Dao`: `First/List/Create/Update/Delete`; filter with `Dao.Query("col = ?", v)`.
- `rdb.Query.Limit(n)` requires `n > 0`; `n == 0` panics, so guard it when building paging.

### 4. Error codes (plain style uses panic)
- In the default plain style, impl/core methods do not return errors; business errors are
  expressed with `ex.PanicNew(ex.ValidationFailed, "...")` and recovered by the framework
  into structured error responses.
- `ex.ValidationFailed` / `ex.NotFound` are **ApplicationError**; `ex.InvalidRequest` is a
  **SystemError**. The plain client re-panics on any non-nil error via `ex.PanicIfError`
  (including ApplicationError), so tests assert with `recover`; only when switching to an ER
  client do you inspect the returned `ex.Error`. Use `ex.ValidationFailed` for business
  validation (e.g. an empty message), never `ex.InvalidRequest`.

### 5. testkit (see app/greeting_test.go)
- `testkit.StartStandalone` is a **process-level singleton**: start it once per test package
  and share it via a top-level test with `t.Run` subtests.
- Tests use `test-hub.yaml` (appConfigs only, no portal) to avoid binding real portal ports.
- `core/auth_service_test.go` verifies that `GreetingActorAuthService` accepts a token issued
  by login and rejects an invalid token with `Unauthorized`; testkit covers the post-admission
  actor-info path.

### 6. Actor authentication (login + token validation + actor consumption)
The intended authentication chain is carried by Vine's actor mechanism and is **pure Go
server-side**:

1. **Contract**: `GreetingActor` declares `auth { credential { token: string } info { username: string } }`.
   skelc generates a `GreetingActorAuthService.Auth(credential) info` (the actor's auth service).
2. **Login** (`AuthService.login`, `noauth`): validates `admin`/`admin123`, then `core/session.go`
   `SessionStore` issues a random token (token→username) and returns `LoginResult{success, username, token}`.
3. **Portal authentication**: on an `auth` method, RPCGW parses the credential from the
   Authorization header, calls `greeting.GreetingActorAuthService.Auth` to validate the token
   (see `impl/greeting_actor_auth_service.go`); on success it injects an **authenticated actor
   + info** into the backend; otherwise the request is rejected.
4. **Welcome consumes the post-admission actor**: `WelcomeService.getWelcome` is `auth` by
   default (no `noauth`). The impl injects `meta.Context`; core reads the username with
   `meta.GetActorInfo[*skeled.GreetingActorInfo](ctx.Actor())` — panicking with
   `ex.Unauthorized` if absent. See `core/welcome_service.go`.

Gotchas:
- **`AuthService.login` must be `noauth`**, otherwise login itself requires auth (deadlock).
- **Methods that need login must NOT write `noauth`** (the default is auth mode).
- The actor's `InfoType` is registered as a **pointer** `*GreetingActorInfo`; read and
  construct with pointer types (`meta.GetActorInfo[*...]` / `meta.NewAuthenticatedActor(&...)`).
- **Demo only**: tokens live in memory (lost on restart) and credentials are hard-coded. A
  real project should persist sessions and use full Portal admission auth
  (actor/auth/credential), not hard-coded accounts.
- Unit tests cover token issuance/validation, and testkit covers business behavior with an
  authenticated actor. Only a real Portal/RPCGW call proves Authorization credential parsing
  and admission wiring.

### 7. Seed / Portal (see seed/hub.yaml)
- The business entry uses **RPCGW** (vRPC over HTTP) with `pathPrefix: /api`; `actorSkelName`
  must exactly equal the generated full Skel name.
- The portal **entry listen port equals the DashboardURL port** (7299 here, the control plane);
  the business entry port is in `portalRules` (7188 here). These are separate portal entries.
- No WEBGW is needed or configured (that is a frontend application's gateway, not part of
  this server template).

## Startup

```bash
# 1. Start the server (from the server Go module root)
cd src/server
go run ./cmd/greeting

# 2. Required delivery check with a vRPC client against Portal RPCGW
#    Entry http://127.0.0.1:7188/api/invoke
#    Login: greeting.AuthService/login (noauth, no Authorization needed)
#    Welcome: greeting.WelcomeService/getWelcome
#        needs Authorization: token <login token> (actor credential)
#    CRUD: greeting.GreetingService/createGreeting etc.
```

Internal behavior is covered by `go test ./...` from `src/server/`. This template does not bundle a browser
client, but before claiming runnable delivery, public Rpc schema availability, or successful
actor admission, call the Portal/RPCGW path above with an actual vRPC client such as
`@yorun-ai/vrpc` and retain the evidence. The Dashboard control plane is at
http://127.0.0.1:7299/ (not the business entry).
