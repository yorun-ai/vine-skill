# Server Example: Greeting

A server-only example for Vine `v0.32.5`, skelc `v0.34.0` and Go `1.27.3` (module floor `1.27.0`). It demonstrates generated API Servers, token-based Actor admission, direct generated config injection and SQLite persistence. Add a requested browser application under `src/web/`, following [frontend-vrpc.md](../frontend-vrpc.md).

The checked-in files are hand-maintained contracts and Go source. Generate the modules before building; validation creates `skeled/`, modules, databases and build output in a disposable temporary directory.

## Generate and Build

From a copy of this directory, with the pinned compiler already installed:

```bash
skelc check --skel-in ./skel
skelc gen go-module --skel-in ./skel --go-out ./skeled/golang \
  --go-module example.com/greeting/skeled/golang --go-vine-version v0.32.5
(cd skeled/golang && go mod tidy)
(cd src/server && go mod init example.com/greeting/src/server)
(cd src/server && go mod edit -go=1.27.0 \
  -require=example.com/greeting/skeled/golang@v0.0.0 \
  -replace=example.com/greeting/skeled/golang=../../skeled/golang)
(cd src/server && go get go.yorun.ai/vine@v0.32.5 && go mod tidy)
(cd src/server && go build ./... && go test ./... && go vet ./...)
```

To validate the maintained package in an isolated directory, run `bash scripts/validate_example.sh` from the repository root. It also generates and compiles external Go/TypeScript API clients, checks the actual dependency graph, and runs Portal/Admin/Web smoke checks. Go downloads, Node/npm and the pinned skelc must be available.

## Contract and Application Boundaries

- `domain.skel` owns the canonical domain declaration; other files use a bare domain header.
- `AuthApiService` and `GreetingApiService` explicitly use `auth optional`; `WelcomeApiService` uses `auth required`. `for GreetingActor` selects the actor audience.
- Backend generation emits API Servers. Generate their external clients separately with `--api` and call them through Portal. `testkit.Execution.NewClient` serves backend `pub service` contracts.
- Adapters in `src/server/impl` embed generated `Default...Server` types and call business services in `core`. Business validation raises `ex.ValidationFailed`; protocol request failures use the SystemError `InvalidRequest`.
- `MainDatabase` owns the SQLite component. `Migrator.BeforeAppStart` explicitly creates the demo table. A concrete `EnsureSchema` implementation is another supported location for application-owned migrations.
- Scalars use `go.yorun.ai/skel/types`; UUID construction uses Go 1.27's standard `uuid` package.
- `GreetingConfig` is injected directly. The server module uses a local replace for the independently generated backend module.

## Authentication and Testing

1. Anonymous `AuthApiService.login` accepts demo `admin` / `admin123` credentials and returns a random session token.
2. Send `Authorization: token <token>` to `WelcomeApiService.getWelcome`.
3. Portal invokes generated `GreetingActorAuthService` for credential validation and supplies authenticated Actor info to the API Server.
4. Optional endpoints admit anonymous calls with omitted credentials and authenticate supplied credentials. Supply a valid token or omit the credential when logging in.

Credentials, token results and Actor credentials carry sensitive markers where appropriate. The demo session store is memory-only and the demo account is fixed; replace these with the target application's authentication/session design for production.

`core/auth_service_test.go` checks token issuance and credential rejection. `app/greeting_test.go` starts one testkit runtime and copies the real Hub seed onto an isolated loopback port. It verifies Portal login/admission, Actor propagation, CRUD persistence and `VALIDATION_FAILED` → HTTP 422. Each run uses temporary storage. Real Portal/RPCGW calls verify admission; seeded test Actors support controlled internal execution tests.

## Seed and Startup

`src/server/seed/hub.yaml` declares a named loopback Portal entry, an RPCGW site and a rule referencing `entryName`. The current rule uses `matchPathPrefix: /api` and `routeSiteName`.

```bash
# From the server Go module root, after checking both ports are available:
cd src/server
go run ./cmd/greeting
```

- Business vRPC: `http://127.0.0.1:7188/api/invoke`.
- Hub Dashboard/Admin: `http://127.0.0.1:7299/` (unauthenticated, loopback only).
- Hub state: `greeting-hub.sqlite`; business state: `greeting.sqlite`.

Standalone uses `HubDBSQLiteFile`, `HubSeedDataFile` and `HubAdminListen`. Persistent Hub seed applies only on first initialization. Preserve stored data and apply later edits through the supported management API or an explicit migration. A new no-db project may omit the DB option and load read-only configuration from seed on each start.

## Required Delivery Check

Generate external clients with the pinned compiler:

```bash
skelc gen ts --api --actor greeting.GreetingActor \
  --skel-in ./skel --ts-out ./skeled/typescript
```

Configure `@yorun-ai/vrpc` `0.10.0` with the business prefix, instantiate `createAuthApiService`, `createGreetingApiService` and `createWelcomeApiService`, and call their generated methods. Login takes `{username, password}`; welcome takes `null` and a per-call Authorization header. Retain success and rejected-credential evidence. The package runner covers this path with a generated TypeScript client running in Node.js.

These tests cover local Portal HTTP and in-process runtime composition. Separated App-to-Link transport, persistent legacy database migration, external MQ durability, multi-host security and browser-session cache behavior require dedicated tests.
