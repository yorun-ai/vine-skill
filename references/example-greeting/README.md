# Server Example: Greeting

A server-only example for Vine `v0.32.5`, skelc `v0.34.0` and Go `1.27.3` (module floor `1.27.0`). It demonstrates generated API Servers, token-based Actor admission, direct generated config injection, SQLite persistence and an Event → Task receipt workflow. Add a requested browser application under `src/web/`, following [frontend-vrpc.md](../frontend-vrpc.md).

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
- `MainDatabase` owns the SQLite component. `Migrator.BeforeAppStart` explicitly creates the greeting and receipt tables. A concrete `EnsureSchema` implementation is another supported location for application-owned migrations.
- Scalars use `go.yorun.ai/skel/types`; UUID construction uses Go 1.27's standard `uuid` package.
- `GreetingConfig` is injected directly. The server module uses a local replace for the independently generated backend module.

## Authentication and Testing

1. Anonymous `AuthApiService.login` accepts demo `admin` / `admin123` credentials and returns a random session token.
2. Send `Authorization: token <token>` to `WelcomeApiService.getWelcome`.
3. Portal invokes generated `GreetingActorAuthService` for credential validation and supplies authenticated Actor info to the API Server.
4. Optional endpoints admit anonymous calls with omitted credentials and authenticate supplied credentials. Supply a valid token or omit the credential when logging in.

Credentials, token results and Actor credentials carry sensitive markers where appropriate. The demo session store is memory-only and the demo account is fixed; replace these with the target application's authentication/session design for production.

`core/auth_service_test.go` checks token issuance and credential rejection. `app/greeting_test.go` starts one testkit runtime and copies the real Hub seed onto an isolated loopback port. It verifies Portal login/admission, Actor propagation, CRUD persistence, `VALIDATION_FAILED` → HTTP 422 and asynchronous receipts. Each run uses temporary storage. Real Portal/RPCGW calls verify admission; seeded test Actors support controlled internal execution tests.

## Event to Task Receipts

Creating a greeting also starts this asynchronous workflow:

```text
GreetingApiService.createGreeting
  → save Greeting in SQLite
  → emit GreetingCreatedEvent
  → Link / embedded NATS / GreetingCreatedListener
  → launch GreetingReceiptTask.prepare
  → Link / embedded NATS / GreetingReceiptRunner
  → insert GreetingReceipt once per jobId
GreetingApiService.getGreetingReceipt → read the stored result
```

Read the implementation in this order:

| Step | Source and generated boundary |
| --- | --- |
| Declare | [greeting_async.skel](skel/greeting_async.skel) defines the Event payload, Task trigger and result; [greeting_service.skel](skel/greeting_service.skel) exposes result lookup. |
| Generate | Run the commands above. Inspect `event.go`, `task.go` and `service.go` in `skeled/golang/` for the emitted signatures. |
| Register | [app.go](src/server/app/app.go) embeds `EventerEnabled` and `TaskerEnabled`, registers the Listener/Runner, and sets each attempt's timeout to 20 seconds and concurrency to 2. Default retries remain enabled. |
| Publish | [core/greeting_service.go](src/server/core/greeting_service.go) injects `GreetingCreatedEventEmitter` and calls `EmitGreetingCreated` after saving the greeting. |
| Consume and launch | [greeting_created_listener.go](src/server/impl/greeting_created_listener.go) embeds `DefaultGreetingCreatedEventListener`, implements `OnGreetingCreated`, and injects `GreetingReceiptTaskLauncher` to call `LaunchPrepare`. |
| Execute | [greeting_receipt_runner.go](src/server/impl/greeting_receipt_runner.go) embeds `DefaultGreetingReceiptTaskRunner` and implements `RunPrepare`. |
| Store and query | [greeting_receipt_repository.go](src/server/repo/greeting_receipt_repository.go) atomically inserts the receipt with a unique job ID. `getGreetingReceipt` returns the stored receipt or `null` while it is unavailable. |

The generated emitter and launcher are injected into their current execution. Vine supplies their context, trace and Link connection. API, Event and Task handling each run in their own execution; asynchronous identity rules are described in [capabilities.md](../capabilities.md#event-and-task).

This example has one creation event per greeting, so the persisted greeting ID is also its stable `eventId`. The Listener uses that `eventId` as `jobId` on every delivery. A database unique constraint plus `ON CONFLICT DO NOTHING` preserves the first receipt and preparation time when an Event or Task is delivered again. Events with multiple occurrences per entity should each carry their own stable occurrence ID.

The create response confirms the greeting write and accepted Event publication. Poll `getGreetingReceipt` with a bounded deadline for asynchronous completion. Using the generated TypeScript service from the delivery check below:

```typescript
const created = await greetings.createGreeting({ message: 'Hello asynchronously' });
const deadline = Date.now() + 10_000;
let receipt = await greetings.getGreetingReceipt({ greetingId: created.id });
while (receipt === null && Date.now() < deadline) {
  await new Promise(resolve => setTimeout(resolve, 100));
  receipt = await greetings.getGreetingReceipt({ greetingId: created.id });
}
if (receipt === null) throw new Error('Receipt is still pending; query again later');
console.log(receipt.message, receipt.preparedAt);
```

The demonstration saves the greeting and publishes its Event in separate operations. A failure between them can leave a saved greeting without a receipt; publication errors reach the caller after the write has committed. Applications requiring recovery across that boundary should write an outbox row in the greeting transaction and publish it with the same stable ID. Embedded NATS uses memory streams; restart recovery requires a durable business record/outbox and appropriately configured transport storage. SQLite retains receipts that have already committed.

The test adapters in [greeting_async_test.go](src/server/app/greeting_async_test.go) deliberately fail once after the Listener launches a Task and once after the Runner commits a receipt. The integration suite waits for Event redelivery and Task retry, verifies the result through Portal, and checks that SQLite contains exactly one receipt for the job. The production handlers are called by those adapters; injected failures live in test files. The generated TypeScript smoke check also exercises the production App's complete success path.

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

These tests cover local Portal HTTP, embedded NATS Event/Task delivery and retries, and in-process App/Link composition. Multi-App fan-out, replica competition, cancellation/timeouts, separated App-to-Link transport, persistent legacy database migration, external MQ durability, multi-host security and browser-session cache behavior require dedicated tests. Cron configuration is described in [capabilities.md](../capabilities.md#cron).
