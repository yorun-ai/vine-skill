# Vine Foundations, Versions, and Application Model

Reviewed 2026-10-26 against released Vine `v0.32.5` and skelc `v0.34.0`. Existing project pins take precedence. For an older project, read [baseline-upgrade.md](baseline-upgrade.md) before changing contracts or persistent Hub data.

## Version-First Preflight

```bash
go version
go env GOWORK
go list -m -json go.yorun.ai/vine go.yorun.ai/skel go.yorun.ai/vrpc
vine version --json
skelc version
```

Inspect `go.mod`, `go.sum`, `go.work`, `replace`, CI installation and actual PATH binaries. Generated descriptors record the compiler revision. Check both declared and selected module versions; using local source requires an active workspace or replace mapping. `vine version --json` is supported at this baseline; inspect `--help` before using it with older binaries.

## Version Baseline Quick Reference

| Item | Reviewed baseline | Evidence |
| --- | --- | --- |
| Go language/module floor | `1.27.0` | Vine/Skel/vRPC Go `go.mod` |
| Tested Go toolchain | `1.27.3` | CI and example validation |
| Vine | `v0.32.5` | Released tag, explicit generation argument |
| Tested skelc | `v0.34.0` | Compiler release and generated descriptor |
| Vine historical compiler floor | `v0.17.1` | `core/skel.MinSkelcVersion()` |
| skelc backend Vine minimum/default | `v0.32.0` / `v0.32.1` | `skelc version`; explicitly override to `v0.32.5` here |
| Selected Skel runtime | `v0.34.0` | Generated module requirement and Go MVS |
| vRPC Go | `v0.15.1` | Vine dependency; external Go API generation pin |
| vRPC TypeScript | `0.10.0` | Generated TS client validation |
| Portal Rpc timeout | 30s default; rejects requests over 120s | RPCGW options and tests |
| Portal Web timeout | Site/path policy, otherwise ordinary HTTP 30s / request max 120s | WebGW policy and tests |
| SSE/WebSocket with default stream budget | Connection lifetime with a 60s traffic-idle limit | Portal stream handling; SSE idle is enforced only at Portal |
| Portal HTTP guardrails | Header read 10s, idle connection 2m, headers 128 KiB; read/write budgets belong to gateway execution | Portal entry server |
| Event/Task attempt defaults | 30s timeout, concurrency 10, retry enabled | Listener/Runner options |
| Retry/durability | Retries continue until acknowledgement; failure storage is application-owned; embedded streams are memory-backed | Link dispatcher and Hub MQ |
| Redis lock lease | 30s with refresh; fencing requires application-owned enforcement | Redis lock implementation |

Verify compatibility for each exact compiler/runtime set and pin the tested versions. Backend output and external API output have different runtime dependencies. The selected Skel version is newer than Vine's declared v0.33.0 requirement; see the upgrade reference for why this is compatible.

## Contracts and Generated Code

Use these commands for a new project after writing `skel/` contracts (replace the module path):

```bash
skelc check --skel-in ./skel
skelc gen go-module --skel-in ./skel --go-out ./skeled/golang \
  --go-module example.com/demo/skeled/golang --go-vine-version v0.32.5
(cd skeled/golang && go mod tidy)
mkdir -p src/server
(cd src/server && go mod init example.com/demo/src/server)
(cd src/server && go mod edit -go=1.27.0 \
  -require=example.com/demo/skeled/golang@v0.0.0 \
  -replace=example.com/demo/skeled/golang=../../skeled/golang)
(cd src/server && go get go.yorun.ai/vine@v0.32.5)
# After implementing imports:
(cd src/server && go mod tidy)
```

For a maintained browser client target:

```bash
skelc gen ts --api --actor demo.ClientActor \
  --skel-in ./skel --ts-out ./skeled/typescript
```

For an external Go client, generate a separate API module:

```bash
skelc gen go-module --api --actor demo.ClientActor \
  --skel-in ./skel --go-out ./skeled/go-api \
  --go-module example.com/demo/skeled/go-api --go-vrpc-version v0.15.1
```

Use the actual full actor name. Preserve existing generation entrypoints and output ownership. Imported domains need `--skel-import domain=PATH` plus the corresponding `--go-import`/`--ts-import` mapping for generated consumers. `schema get` resolves imports by default. Choose resolved inspection with import mappings, or `--shallow` inspection with opaque imported references.

For public/regular Go package splitting and cross-domain calls, use [Orders and Catalog](example-domains/README.md#contracts-and-generation). A `pub service` exports its Client and a `pub event` exports its Listener; the owning regular package supplies the Server and Emitter. Generate both sides together with `--go-pub-out` and keep each generator's output directory separate.

`go-module` owns generated module metadata as well as Go source. `gen go` is the non-module variant. Keep hand-written code outside generated directories. New output registers `descriptor.Domain` and imports scalars from `go.yorun.ai/skel/types`. Use standard-library `uuid` when constructing `skeltype.UUID`; wire JSON remains a UUID string.

### Backend Services and APIs

- `pub service OrderService`: Clients and Servers for trusted backend Rpc; actor audiences and admission belong to API/Web contracts.
- `api service OrderApiService`: external API Servers; generate separate Go/TS API clients with `--api`.
- API declarations require an explicit authentication mode and actor audience. Select actors with `for ClientActor`.
- `ext service` exports a Server for another domain to implement; `ext event` exports an Emitter for another domain to publish. Declare Portal API exposure explicitly with an API contract.
- Name Services after an entity or use case. An App can register several handlers sharing a business implementation.

## New-Project Standard Structure

Vine supports project-defined layouts. This Skill's default for a new project is:

```text
demo/
├── skel/                        # hand-maintained contracts
├── skeled/
│   ├── golang/                  # generated backend Go module
│   └── typescript/              # when a TS consumer is needed
└── src/
    ├── server/                  # server Go module
    │   ├── app/app.go           # assembly
    │   ├── cmd/demo/main.go     # process entry
    │   ├── core/                # business behavior and interfaces
    │   ├── impl/                # generated Server adapters
    │   ├── repo/                # persistence
    │   └── seed/hub.yaml        # Hub configuration
    └── web/                     # only for a requested/confirmed browser UI
```

Keep `src/server/app/app.go`, `src/server/cmd/demo/main.go`, `src/server/impl/`, `src/server/repo/` and `src/server/seed/hub.yaml` traceable to their responsibilities. `skeled/golang/go.mod` is separate from `src/server/go.mod`; the latter contains:

```go.mod
require example.com/demo/skeled/golang v0.0.0
replace example.com/demo/skeled/golang => ../../skeled/golang
```

Preserve existing layouts; migrate them when requested. Create directories as their responsibilities arise, and add a frontend when requested.

## Application Construction and Lifecycle

```go
standalone.NewWithOption[*GreetingApp](standalone.Option{
    HubSeedDataFile: "./seed/hub.yaml",
    HubAdminListen:  "127.0.0.1:7299",
}).StartAndWait()
```

This example selects read-only, seed-backed no-db mode. For writable persistence set `HubDBSQLiteFile` or `HubDBPostgresURL`. The Dashboard/Admin listener is independent of Portal and relies on loopback isolation or an authenticated management boundary for access control. Set its address explicitly when needed, and check ownership before starting.

`New` constructs and injects the App spec, calls `DIInit`, validates its identity and captures listener inputs. Components and Modules are constructed during `Start`. Put startup checks before public registration, and background loops in hooks that retain cancellation and join handles. `BindCommon`/`Bind` may run in several containers: declare bindings only.

- Startup: App `BeforeAppStart` hooks → Components then Modules before hooks → endpoint/capability start and registration → Component/Module after hooks → App after hooks.
- Shutdown: App before hooks → Modules then Components before hooks in reverse order → unregister/drain/server stop → Module/Component after hooks in reverse order → App after hooks.
- `InitHooks` registers injected callbacks directly on the App. Check the target API for hook order when extending existing assembly.
- Requests may arrive before `AfterAppStart` returns. Before hooks perform bounded readiness; after hooks must tolerate visible traffic.
- `BeforeAppStop` stops producers and joins workers; `AfterAppStop` releases resources needed by drained requests. Implement cleanup and rollback explicitly for resources acquired before a hook failure.

App names accept digits within lowercase dotted segments. Application identity is distinct from executable build metadata: use injected `meta.CurrentApp` for the former and `buildinfo` for the latter.

## Configuration Lifecycle

Inject generated `*skeled.NameConfig` directly.

| Lifecycle | Behavior |
| --- | --- |
| `eternal` | Freezes the instance snapshot after its first resolution |
| `instant` | Later resolutions decode a new pointer from the current snapshot; existing pointers are unchanged |

Singletons holding an `instant` pointer keep that original value. Resolve refreshed config in a new execution or use an explicit refresh design. Values require a registered concrete config and valid data; resolution failures raise errors. Preserve string whitespace unless application logic explicitly trims it.

Structured fields, nullable values and abstract adoption are supported:

```skel
abs config DatabaseAbsConfig eternal {
    endpoint: string
}
pub config MainDatabaseConfig = DatabaseAbsConfig
```

Put declarations in a file with the appropriate domain header. Abstract configs define reusable contracts; concrete adopters own stored values. Concrete adopters have independent identities, inherit lifecycle/fields/sensitivity, and generate `AsDatabaseAbsConfig()` conversions. Register and seed the concrete full name. Use the target's explicit matching/binding rules to select one concrete adopter for abstract DI.

Hub accepts structured YAML config values or JSON strings. No-db mode loads the seed on each fresh start. SQLite/PostgreSQL imports seed during initial setup, then stored data is authoritative. Update stored configuration through the management API or a reviewed migration.

## Public Package Boundaries

| Purpose | Packages |
| --- | --- |
| App and tests | `app`, `app/standalone`, `app/linked`, `app/testkit` |
| Capabilities | `core/conf`, `core/rpc`, `core/web`, `core/event`, `core/task`, `core/skel` |
| Execution | `core/di`, `core/ctr`, `core/meta`, `core/ex`, `core/logger`, `core/redact` |
| Runtime coordination | `core/lock` |
| Build identity | `buildinfo` |
| Data | `infra/rdb`, `infra/redis` |
| Shared contracts/wire | `go.yorun.ai/skel/types`, `go.yorun.ai/skel/descriptor`, `go.yorun.ai/vrpc/transport/http` |

Use internal source as research evidence and the public packages above as application dependencies. For exact APIs and release evidence see [official-sources.md](official-sources.md).
