# Vine Foundations, Versions, and Application Model

This reference is distilled from the Vine `v0.12.0` source and the website's `next` documentation as of 2026-08-03. Confirm the target project's pinned version first. When versions differ, use the target revision's source, GoDoc, CHANGELOG, and CLI `--help`.

## Contents

- [Version-first preflight](#version-first-preflight)
- [Runtime mental model](#runtime-mental-model)
- [Contracts and generated code](#contracts-and-generated-code)
- [Wire a minimal app after generation](#wire-a-minimal-app-after-generation)
- [ApplicationSpec and runtime modes](#applicationspec-and-runtime-modes)
- [Component, Module, and lifecycle](#component-module-and-lifecycle)
- [Configuration lifecycle](#configuration-lifecycle)
- [Recommended project structure](#recommended-project-structure)
- [Public package boundaries](#public-package-boundaries)
- [Compatibility upgrade checklist](#compatibility-upgrade-checklist)

## Version-First Preflight

A build is determined jointly by Go, `go.yorun.ai/vine`, and skelc. Check these in the target module first:

```bash
go version
go list -m -json go.yorun.ai/vine
vine version --json
skelc version
```

Then inspect:

- The `go`, `toolchain`, and Vine revision in `go.mod`.
- CI images, tool installation scripts, and the actual `PATH`.
- `.skel` files, generated directories, and generation commands.
- The skelc version recorded by generated schemas.
- Upgrade notes in `CHANGELOG` from the project's version to the target version.

For the `v0.12.0` and current `next` baseline:

| Item | Baseline |
| --- | --- |
| Go | `1.26.5` or later |
| Vine | `v0.12.0` source baseline |
| Minimum skelc | `v0.9.0` |

The minimum skelc version is only the runtime floor, not a recommendation to use the latest version. Future generators have no promised upper compatibility bound. For production, use an exact tag or commit that has been used to generate, review, and test the application.

Do not leave `@main` or `@latest` in production automation. Commit `go.mod`, `go.sum`, `.skel`, and generated code. Before 1.0, the website maintains only `next`, which may be ahead of the latest release.

## Runtime Mental Model

```text
Skel contract ──skelc──> generated Go/TypeScript boundary
                              │
                              ▼
Hub (control-plane state) ──> Link (app ingress, discovery, delivery) <──> App (business execution)
          │                    ▲
          └────────────> Portal (external HTTP/HTTPS entry)
```

- App owns Components, Modules, Handlers, Listeners, Runners, and business state.
- Link handles App registration, configuration reads, service discovery, forwarding, Event/Task consumers, and draining.
- Hub manages configuration, schemas, registrations, leases, Portal configuration, and runtime distribution. It is not on every synchronous request path.
- Portal handles only external HTTP/HTTPS entry and policy. App-to-App calls do not pass through Portal.
- Standalone preserves the same responsibilities but replaces some network transports with in-process transports.

Collaboration inside one App normally uses plain Go DI. Use Skel Rpc/vRPC for every structured synchronous business call across a capability boundary, including calls across Apps, from browsers, and from external clients. Do not replace clear in-process dependencies with network contracts, and do not replace Rpc with Web/HTTP JSON routes.

## Contracts and Generated Code

Pin the revision first, then initialize only an empty project, Go module, `skel/`, generated directory, and repeatable generation entry. These are generation prerequisites, not business implementation:

```bash
mkdir vine-hello
cd vine-hello
go mod init example.com/vine-hello
go get go.yorun.ai/vine@"$VINE_REVISION"
mkdir skel skeled
```

Standard pipeline:

```text
.skel source → skelc check → skelc gen → generated Go/TypeScript types/Schema/Server/Client → business implementation
```

This is a contract and generated-boundary gate, not a mandatory ritual for ordinary implementation changes. When adding or modifying `.skel`, public wire shapes, capability registration, Vine/skelc versions, or generation commands, do not write Handlers, Servers, DTOs, routes, or client adapters that depend on the new boundary until `skelc check`, pinned-version generation, and generated-boundary review all succeed.

If a change only modifies business algorithms behind an existing contract, DAOs, logging, or internal lifecycle behavior, without touching the contract, capability registration, versions, or generation commands, leave `.skel` and generated directories unchanged and run the target build and behavior tests. If generated code is missing, stale, or inconsistent with the schema, stop immediately and re-enter the full generation gate.

Never bypass Skel by hand-writing parallel interfaces or modifying generated files.

Minimal commands:

```bash
skelc check --skel-in ./skel
skelc symbol list --skel-in ./skel
skelc gen go --skel-in ./skel --go-out ./skeled
```

The last command demonstrates only the Go target. When the repository maintains a browser client, its existing pinned generation entry must also produce the TypeScript target in the same gate. Do not guess TypeScript subcommands or flags from another skelc revision; use the repository script and the pinned tool's `--help`.

Prefer the repository's existing generation script, Make target, or `go generate` so that parameters, import rewriting, and formatting stay consistent. Regeneration overwrites generated directories:

- Modify `.skel`, not generated Go or TypeScript.
- When implementing a Server outside the generated package, embed the generated `Default...Server`; the generated interface may contain a private sealing method and cannot be implemented from scratch.
- Go callers should use generated clients; browser callers should use generated TypeScript services/specs/types through the project vRPC client. Event and Task code should use generated emitters, launchers, Listeners, and Runners.
- After changing Vine or skelc, regenerate every maintained contract and language target and review the diff.

Skel syntax and skelc commands belong to the Skel documentation. Do not guess language rules in the Vine skill.

## Wire a Minimal App After Generation

Enter this section only after the previous Skel checks, generation, and generated-boundary review succeed. The following example shows only the App lifecycle shell after generation. Real capabilities must wire generated Rpc/Event/Task, or generated Web only when binary-stream semantics require it:

```go
package main

import (
    "go.yorun.ai/vine/app"
    "go.yorun.ai/vine/app/standalone"
    "go.yorun.ai/vine/core/logger"
)

type HelloModule struct {
    app.BaseModule
}

func (*HelloModule) AfterAppStart() {
    logger.Info("hello from Vine")
}

type HelloApp struct {
    app.Application
}

func (*HelloApp) Name() string {
    return "demo.hello"
}

func (*HelloApp) InitModules(add app.TypeAdder) {
    add(app.T[*HelloModule]())
}

func main() {
    standalone.NewWithOption[*HelloApp](standalone.Option{
        SQLiteFile: "./vine.sqlite",
    }).StartAndWait()
}
```

`SQLiteFile` is the embedded Hub database, not the business RDB. `StartAndWait()` waits for SIGINT/SIGTERM and stops gracefully. App names must match `^[a-z]+(?:\.[a-z]+)*$`.

## ApplicationSpec and Runtime Modes

`ApplicationSpec` declares:

| Entry | Responsibility |
| --- | --- |
| `Name()` | Stable logical application name |
| `InitComponents` | Infrastructure Components owned by the application |
| `InitModules` | Business Modules |
| `BindCommon` | Application-wide shared dependencies |
| `ServicerInitHandlers` | Rpc service implementations |
| `WebberInitHandlers` | Web Handlers |
| `EventerInitListeners` | Event Listeners |
| `TaskerInitRunners` | Task Runners |

Declare only the capabilities the App actually needs. Undeclared capabilities are not included in Link registration.

Runtime modes:

```go
// Single-process development and integration tests
standalone.NewWithOption[*DemoApp](standalone.Option{
    SQLiteFile: "./vine.sqlite",
}).StartAndWait()

// In-process Link connected to an external Hub
linked.NewWithOption[*DemoApp](linked.Option{
    HubEndpoint: "http://127.0.0.1:7071",
}).StartAndWait()

// Separate App connected to a Link sidecar
app.NewWithOption[*DemoApp](app.Option{
    LinkEndpoint: "http://127.0.0.1:7079",
}).StartAndWait()
```

`New` immediately constructs and injects the spec, calls `DIInit()`, validates the name, derives the root context, and captures listener inputs. It is not a lazy factory. Each spec type and App name can be constructed only once per process, and an App can start and stop only once.

Pass flags at the constructor boundary. When the App provides its own defaults, normalize them in the spec's `DIInit()`. Changing `ListenAddr` in `BindCommon` happens after the value has already been captured.

## Component, Module, and Lifecycle

- Component: application-owned infrastructure such as databases and Redis; exposes connections, DAOs, caches, or lockers to DI.
- Module: domain services, background workers, and business lifecycle resources.
- Both are App-lifecycle singletons. They are wired during `Start()`, not `New()`.
- `BindCommon` and object `Bind` may be applied to multiple containers and must be deterministic and free of side effects.

Startup order:

1. Initialize Link/config readers and the root injector.
2. Construct Components, then Modules.
3. Construct capability servers and execution containers.
4. Run Component and Module `BeforeAppStart` hooks in declaration order.
5. Start endpoints and capabilities, then register with Link.
6. Run Component and Module `AfterAppStart` hooks in declaration order.

Shutdown order:

1. Run Module and Component `BeforeAppStop` hooks in reverse order.
2. Unregister capabilities and drain.
3. Stop servers and cancel the root context.
4. Run Module and Component `AfterAppStop` hooks in reverse order.

Responsibilities:

| Hook | Put here |
| --- | --- |
| `BeforeAppStart` | Bounded dependency checks, warmup, and pre-registration readiness |
| `AfterAppStart` | Background loops that remain safe after the application is visible |
| `BeforeAppStop` | Stop producers, cancel/join workers, and perform bounded flushes |
| `AfterAppStop` | Release final resources owned by the application |

A `BeforeAppStart` error becomes a panic, and startup has no transactional rollback. Hooks have no framework-provided automatic timeout. Declaration order determines hook order, but dependency resolution may construct objects earlier; express dependencies through real DI.

## Configuration Lifecycle

```text
Skel config → Hub DB/seed → Link snapshot → DI factory → typed Go pointer
```

| Lifecycle | Visible behavior | Suitable for |
| --- | --- | --- |
| `eternal` | Freezes an instance snapshot after the App first reads it | Connections, schemas, startup policy |
| `instant` | Later DI resolutions decode a new pointer from the latest Link snapshot | Feature flags, limits, dynamic behavior |

Both are lazy. `instant` does not mutate existing objects: once a Module, Component, or explicit singleton stores a pointer, it continues to see the value from construction time. Resolve behavior that needs the latest value in a new execution, or design an explicit refresh boundary.

Decoding is strict: the generated type must be registered, Hub/Link must contain a non-empty value under the full Skel name, and the JSON must decode successfully. Failures do not return a zero value. Seed YAML's `value` is JSON encoded inside a YAML string; after import, the Hub database remains the source of truth.

Use `eternal` for long-lived invariants. Use `instant` only when each new execution can safely select the new value. Keep related fields backward compatible during rolling deployments.

## Recommended Project Structure

Vine does not require a directory layout. A medium-sized project can use:

```text
demo/
├── cmd/demo/main.go             # Select the runtime mode and start only
├── internal/application/        # App spec and Component/Module wiring
├── internal/account/            # Business Module, services, capability Handlers
├── internal/platform/           # RDB, Redis, and other Components
├── skel/                        # Hand-maintained contracts
├── skeled/                      # Generated code
├── config/                      # Optional application configuration
├── migrations/                  # Explicit database migrations
├── go.mod
└── go.sum
```

Organize source by business capability and keep tests in the same package as the Go files under test. Small projects can be simpler. Large projects should follow the team's existing Go conventions. This tree is guidance, not a framework contract.

## Public Package Boundaries

| Task | Public packages |
| --- | --- |
| App wiring | `app`, `app/standalone`, `app/linked`, `app/testkit` |
| Capabilities | `core/conf`, `core/rpc`, `core/web`, `core/event`, `core/task`, `core/skel` |
| Execution | `core/di`, `core/ctr`, `core/meta`, `core/ex`, `core/logger`, `core/redact` |
| Runtime/build identity | `core/runtime`, `buildinfo` |
| Infrastructure | `infra/rdb`, `infra/redis` |
| Utilities without runtime dependencies | `util/*` |

Use `internal/*` only to understand framework implementation and stack traces; it is not an application API. `core/runtime` describes process and executable identity, not one App inside a bundle.

## Compatibility Upgrade Checklist

1. Change the pinned Go, Vine, and skelc values together on a reviewable branch.
2. Review the target Vine module's Go requirement and every upgrade note.
3. Run the target binary's `vine version --json` and record its minimum skelc.
4. Regenerate every contract with the selected skelc.
5. Review generated diffs; do not patch generated directories by hand.
6. Run application tests and static checks.
7. In staging with the target topology, validate registration, configuration, calls, delivery, draining, and security boundaries.
8. Release the business App and matching runtime services as one compatibility change.

See [official-sources.md](official-sources.md) for website entries and exact source locations.
