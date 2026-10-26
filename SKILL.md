---
name: vine-skill
description: Develop, troubleshoot, review, test, upgrade, or deploy Yorun Vine applications using Go and Skel contracts. Use for Vine application wiring, generated Rpc/API/Web/Event/Task boundaries, configuration, data components, and Hub/Link/Portal integration. Resolve the project's pinned versions before choosing APIs; use read-only diagnosis and perform implementation when requested.
---

# Develop Vine Applications

## Follow the Core Boundaries

- **Pin versions first.** Read the target project's Go modules, toolchain, generation scripts and generated markers. This package is tested with Vine `v0.32.5`, skelc `v0.34.0`, Go `1.27.3`, vRPC Go `v0.15.1` and vRPC TS `0.10.0`. Existing project pins take precedence. Current website documentation can be ahead of a release or retain stale examples; resolve conflicts against the pinned source and CLI.
- **Generate before depending on a changed contract.** Changes to `.skel`, generated boundaries, capability registration or toolchain versions require pinned validation, generation for every maintained target and generated-diff review. Implementation-only changes can preserve the generated boundary.
- **Distinguish backend Rpc from Portal APIs.** `pub service FooService` generates backend Clients and Servers for App-to-App calls. `api service FooApiService` generates Servers for external access through Portal; generate its external Go/TypeScript clients separately with `--api`.
- **Use the matching capability.** Prefer typed Rpc/vRPC for structured business calls. Use Web for files, assets and HTTP/stream protocols that require it. Share in-process business behavior through Go interfaces and DI.
- **Preserve execution scope.** Create, use and release request contexts, clients, DAOs, caches and lockers within their owning execution. Preserve trace, Actor, Initiator, cancellation and remaining deadlines.
- **Preserve user intent.** Review and diagnosis stay read-only. Implement the requested scope, respect existing project conventions and authorization, and distinguish completed checks from validation gaps.
- **Build the server first.** Validate the server, then build a browser UI when the user requests one. An explicit frontend request supplies confirmation. For a server-only request, deliver the server and ask whether a browser UI is needed afterward.

## Select the Reference

Read only the references needed by the task:

| Task | Reference |
| --- | --- |
| Pins, generation, structure, application and configuration | [foundations.md](references/foundations.md) |
| Upgrade an older Vine project | [baseline-upgrade.md](references/baseline-upgrade.md) |
| Rpc/API/Web/Event/Task and admission | [capabilities.md](references/capabilities.md) |
| DI, context, filters and errors | [execution-and-boundaries.md](references/execution-and-boundaries.md) |
| RDB, Redis, runtime locks, logging and testing | [data-and-testing.md](references/data-and-testing.md) |
| Standalone/linked/separated, seed and deployment | [runtime-operations.md](references/runtime-operations.md) |
| Requested browser UI, generated TS and Web serving | [frontend-vrpc.md](references/frontend-vrpc.md) |
| Exact source and release evidence | [official-sources.md](references/official-sources.md) |
| Starting a server implementation, including Event → Task delivery | [example-greeting](references/example-greeting/README.md) |
| Cross-domain public contracts, two-App Rpc calls and Events | [example-domains](references/example-domains/README.md) |

## Apply the Generation Gate by Impact

1. Identify the selected compiler binary and actual dependency graph, including `go.work` and `replace`. Record each dependency's declared requirement and the version selected for the application.
2. Update the hand-maintained contract when needed. Run the pinned `skelc check`; provide dependency mappings to generation and schema inspection when imports exist.
3. Generate the backend and every maintained external client target. Review the public names, authentication policies, descriptors, scalar imports and module requirements.
4. Then implement consumers. Embed generated `Default...Server` types; use generated clients, Listeners, Runners, emitters and configuration types. Make contract changes in Skel and regenerate the output.
5. Compile the actual application graph and test the affected runtime boundary. Select consumer upgrades by generated/runtime contract impact; retain compatible consumer pins and record the evidence.

`skelc v0.34.0` uses `go.yorun.ai/skel`, generates `descriptor.go`, and requires Vine `v0.32.0+` for backend output. Its `--go-vine-version` default is `v0.32.1`; explicitly pass `v0.32.5` to reproduce this package's tested baseline. See the foundations reference for commands.

## Build and Wire the Application

For a new project's default layout, use root `skel/`, generated `skeled/golang/` and, when needed, `skeled/typescript/`, plus hand-written `src/server/`. Put assembly in `src/server/app/app.go`, entry points in `src/server/cmd/<name>/main.go`, and the seed in `src/server/seed/hub.yaml`. The server module requires the generated module at `v0.0.0` with `replace <module>/skeled/golang => ../../skeled/golang`. Add a confirmed browser package under `src/web/`. Preserve existing layouts; perform layout migrations when requested.

Split Services by entity, aggregate or named use case. An App owns runtime assembly. Give each business layer a concrete responsibility, such as validation, orchestration or persistence. Inject generated `*skeled.<Name>Config` directly. Abstract config adoption is a separate supported contract, described in the foundations reference.

Components own infrastructure; Modules own business lifecycle resources. Use `BindCommon` and `Bind` for deterministic binding declarations; perform resource startup in lifecycle hooks. Listener inputs belong at construction or `DIInit`, before they are captured. App names accept lowercase segments and digits: `^[a-z][a-z0-9]*(?:\.[a-z][a-z0-9]*)*$`.

Use public `app`, `core/*`, `infra/*`, `util/*`, `buildinfo`, and the generated facades. Use `go.yorun.ai/skel/types` for scalars and standard-library `uuid` with Go 1.27. Application identity comes from injected `meta.CurrentApp`; executable metadata comes from `buildinfo`.

## Keep Runtime Boundaries Explicit

- Configure Standalone through `HubSeedDataFile`. Set loopback `HubAdminListen` to enable Dashboard and Admin on Hub. This listener relies on loopback isolation or an authenticated management boundary for access control; an empty address disables it.
- Hub defaults to seed-backed, read-only no-db mode. SQLite/PostgreSQL persistence imports seed during initial setup, then uses stored data. Apply subsequent changes through the management API or a reviewed migration.
- Portal requires explicit entries and rules referencing `entryName`. Use the tested seed's `protocol`/`http` entries, `matchPathPrefix`/`routeSiteName` rules and explicit authentication modes.
- Keep Hub, Link, Portal, embedded Dashboard assets and application dependencies on compatible reviewed versions. For a large upgrade, read database migration floors before starting a new binary against persistent data.
- Use Standalone for local composition. Use separated Hub/Portal/Link/App processes to validate wire transport between applications, discovery leases, mTLS and independent failures.
- In-process Rpc copies declared Go values through runtime cloning. Validate value isolation locally and JSON/CBOR wire behavior through an actual transport.
- Use generated API clients through a real Portal to validate external admission. Use Hub Debug and seeded test Actors for controlled internal executions.

## Validate and Deliver

| Change | Required evidence |
| --- | --- |
| Contract/generation or baseline | Pinned check/generation, reviewed output, selected dependency graph, affected Go/TS builds |
| Go behavior | Formatting, focused behavior tests, appropriate package/repository tests and vet |
| Backend Rpc | Generated client and testkit execution, plus separated transport if wire behavior changes |
| API/admission | Real Portal calls: success, relevant auth/error case, trace and status mapping |
| Web/frontend | Compile the generated handler; proxy/assets behavior tests; real WEBGW page and generated vRPC call; frontend checks/build |
| Persistence/config | Fresh initialization and relevant restart/upgrade tests; verify no-db vs persistent behavior |
| Concurrency/lifecycle | Race checks and cancellation, failure, cleanup and duplicate-delivery cases as applicable |
| Deployment/topology | Staging paths and failure drills matching the deployed topology |

Run `git diff --check`. Report versions, changes, checks actually run and material gaps. Tie each conclusion about wire behavior, admission, durability or production readiness to its corresponding test evidence. When delivering a browser app, give its prerequisites, startup/preflight commands for supported platforms, business/API/Dashboard addresses, and any unresolved startup issue. Adapt the optional launcher templates to the project's actual ports and platform requirements.
