---
name: vine-skill
description: Develop, troubleshoot, review, test, upgrade, or deploy Yorun Vine services. Use for Go projects importing go.yorun.ai/vine and work involving Vine apps, components, modules, DI, lifecycle, Skel-generated Rpc/Web/Event/Task/config contracts, Hub, Link, Portal, data infrastructure, compatibility, or production readiness. Enforces pinned Go/Vine/skelc baselines, generation before implementation for contract or capability changes, generated-boundary preservation for implementation-only changes, and server-first delivery with an optional user-confirmed browser frontend.
---

# Develop Vine Applications

## Follow the Core Boundaries

Always follow these principles:

1. **Pin versions first**: Determine the project's pinned Go, Vine, and skelc versions before selecting APIs, CLI flags, or generation commands. Before Vine 1.0, the website's `next` documentation may be ahead of released versions.
2. **Apply the generation gate by impact**: When adding or changing a Skel contract, generated boundary, capability registration, or Vine/skelc version, complete the pinned `skelc check`, code generation, and generated-diff review first. Implementation-only changes must not modify generated files; re-enter the gate immediately if the generation baseline is inconsistent.
3. **Prefer generated vRPC**: Declare structured synchronous business calls as Rpc and invoke them through generated clients or Portal vRPC. Use Web/HTTP only for binary streams such as file uploads, downloads, Range, and multipart, and for static asset delivery.
4. **Separate responsibilities**: Let App declare capabilities, Module own business behavior, Component own infrastructure, Link handle discovery and delivery, Hub manage control-plane state, and Portal admit external traffic.
5. **Keep scopes correct**: Keep request-related contexts, clients, configuration, DAOs, caches, and lockers within the execution that created them.
6. **Be honest about topology**: Treat standalone results only as single-process evidence. Use linked or separated topologies to validate networking, leases, draining, TLS, and independent failures.
7. **Keep the control plane on one revision**: Treat Dashboard assets, Hub schema, and Portal routes as one versioned unit. Do not reuse a Dashboard origin across projects or Vine revisions during local development.
8. **Act only as authorized by the request**: Stay read-only for review or diagnosis. Modify code and run the corresponding validation only when implementation is explicitly requested.
9. **Build the server first**: For a new project, implement and validate `src/server/` first. Do not pre-create `src/web/` or add Vite, frontend dependencies, or frontend startup configuration. If the original request did not mention a frontend, ask the user whether one is needed after the server is complete and wait for the answer.
10. **Add the frontend only after confirmation**: Only when the user already requested or explicitly confirmed a frontend, create the browser package in `src/web/` and connect it to the existing server through the generated TypeScript vRPC client and Portal WEBGW. Do not move `src/server/`. Copy and adapt the native startup scripts only at this stage.
11. **Use the fixed project skeleton**: For new projects, put hand-maintained contracts in root `skel/`, generated artifacts in root `skeled/golang/` and `skeled/typescript/`, the Hub seed fixed in `src/server/seed/hub.yaml`, hand-written server code under `src/server/` (`app/`, `cmd/<name>/`, `core/`, `impl/`, `repo/`), and the browser package in `src/web/`. Make `skeled/golang/` an independent generated-code module and import it from `src/server/go.mod` with `require` plus a local `replace`. The default skeleton must include `src/server/` and no root-level `app/`, `cmd/`, `core/`, `impl/`, `repo/`, `seed/`, `web/`, or `internal/`. Preserve an existing layout unless a migration is requested.
12. **Split Services by entity**: An App is a runtime and assembly boundary, not a business Service. When one App contains multiple entities or aggregates, split them by concrete entity or named use case across `skel/`, `src/server/impl/`, `src/server/core/`, and `src/server/repo/` (e.g. `CustomerService`, `DealService`, `ActivityService`, `DashboardService`). Do not create a single App-named umbrella Service that stuffs every entity's CRUD, rules, and persistence into one struct.
13. **Inject the generated Config directly**: The `*skeled.<Name>Config` generated from Skel config is the configuration type used by application code. Components, Modules, Services, and Handlers should inject it directly with `inject:""`. Do not copy fields into a `core.Config`, and do not use `NewConfig` or a `BindCommon` factory to wrap the generated Config meaninglessly.
14. **Frontend delivery must state the startup method**: Whenever `src/web/` is created or modified, the final reply must explicitly give the required versions, the first startup command for Linux/macOS, subsequent startup commands, the Windows command, the preflight-only command, the business page/vRPC/Dashboard addresses, and how to handle occupied ports or missing dependencies. Saying the code is done is not enough, and referencing the README is not enough.

## Decide Whether the Generation Gate Applies

Enter the generation gate for any of the following:

- Add, remove, or modify `.skel` data types, Actors, errors, Rpc, Web, Event, Task, or config.
- Change public methods, parameters, return values, full Skel names, serialization shapes, or cross-boundary errors.
- Add, remove, or adjust App capability registration so that a schema or capability starts or stops being exposed.
- Change Vine, skelc, generation commands, generation parameters, generated directories, or generated package boundaries.
- Existing generated code is missing, stale, uncompilable, or inconsistent with the pinned contract.

After entering the gate, follow this exact order:

1. Confirm the project's pinned Go, Vine, and skelc versions and its existing generation commands.
2. Modify the hand-maintained `.skel` first; keep source files unchanged when the contract does not need to change.
3. Run `check` with the pinned skelc version.
4. Regenerate code with the project's existing commands.
5. Review the generated diff, schema, Go Server/Client, TypeScript service/spec/types, Listener, Runner, and every affected package boundary.
6. Only after all of the first five steps succeed, modify Handlers, capability wiring, call adapters, and other implementations that depend on the generated boundary.
7. If the contract changes again, return to steps 2–5.

Only tasks covered by this gate are blocked when skelc is unavailable, versions do not match, checks fail, or generation fails. Never fabricate generated types, hand-write parallel DTOs, interfaces, or routes, or promise to add Skel later.

An implementation-only change must satisfy all of these conditions:

- It does not change `.skel`, generated code, the public wire shape, capability registration, Vine/skelc versions, or generation commands.
- It only changes business algorithms behind an existing contract, DAO queries, logging, internal services, lifecycle behavior, or other hand-written behavior.
- It leaves generated directories unchanged and confirms the existing generation baseline through `go.mod`, generated-file markers, and the project build.
- It runs targeted tests matching the impact. If the schema, generated types, or registration are inconsistent, stop and re-enter the generation gate.

## Select the Capability Boundary

| Need | Choice |
| --- | --- |
| Collaboration inside one App without a capability boundary | Plain Go interfaces and DI |
| Structured synchronous business requests/responses, including across Apps, browsers, and external clients | Skel Rpc with generated clients; use vRPC at external entries |
| Multipart, upload, download, Range, `Content-Disposition`, binary streams, or static assets | Web/HTTP |
| Broadcast a fact that has already occurred | Event |
| Ask one available worker to perform work | Task |
| Declarative runtime parameters | Skel config |
| Caching or distributed coordination | Redis Component |
| Relational persistence | RDB Component with schema migrations managed separately |

Do not create JSON REST Web routes for CRUD, queries, search, pagination, authentication, or state changes. When one feature includes both metadata and file bytes, use Rpc/vRPC for the metadata, authorization, and transfer session, then Web/HTTP for the bytes, bound to the same identity and business ID. Choose Rpc when uncertain.

## Establish the Project Baseline First

Before modifying anything, complete these checks:

1. Read every applicable repository-level and directory-level instruction supplied by the active agent host, then read the project README and contribution guidance.
2. Inspect the worktree and protect the user's existing changes.
3. Locate every `go.mod`, `.skel` file, generated directory, App specification, process entry point, and deployment configuration.
   For a new project, create the [standard skeleton](references/foundations.md#new-project-standard-structure) first, then write contracts or implementation.
4. Confirm actual versions in the target Go module:

```bash
go version
go list -m -json go.yorun.ai/vine
vine version
skelc version
```

Run only commands already available locally. Do not install or upgrade tools merely to probe versions. If commands are unavailable, gather evidence from `go.mod`, `go.sum`, CI, generated files, and build scripts, and state what remains unknown.

5. Identify the current runtime mode:
   - `app/standalone`: embeds Hub, Portal, and Link.
   - `app/linked`: runs the business App and Link in one process and connects to an external Hub.
   - `app.New`: connects the business App to a separate Link.
   - `vine dev`: runs a local runtime process connected to a separate business App.
6. Report the version and topology baseline. Do not treat the website's current examples as the project's API automatically.
7. When using the Standalone Dashboard, check whether another process owns the target host/port, and explicitly set a project-specific `DashboardURL`. Do not rely on the default `:7099`, and do not reuse a browser origin that has served another Vine revision.

See [foundations.md](references/foundations.md) for detailed baseline checks; the `v0.13.1` default timeouts, ports, and liveness timings are collected in [Version baseline quick reference](references/foundations.md#version-baseline-quick-reference).

## Load References by Problem

Read only the references needed for the current task. Combine them for cross-domain changes.

| Problem | Read |
| --- | --- |
| Versions, first app, Skel workflow, project structure, App, Component, Module, configuration | [foundations.md](references/foundations.md) |
| Rpc, Web, Event, Task, and vRPC HTTP | [capabilities.md](references/capabilities.md) |
| User-confirmed browser frontend, generated TypeScript vRPC client, and Portal validation | [frontend-vrpc.md](references/frontend-vrpc.md) |
| DI, execution, Filter, Meta, Actor, trace, timeout, and errors | [execution-and-boundaries.md](references/execution-and-boundaries.md) |
| Redis, RDB, logging, redaction, and testkit | [data-and-testing.md](references/data-and-testing.md) |
| Lifecycle, routing, topology, Hub, Link, Portal, CLI, security, rollout, and failure drills | [runtime-operations.md](references/runtime-operations.md) |
| Finding the exact website pages, source, release notes, or refreshing version facts | [official-sources.md](references/official-sources.md) |
| Installing, discovering, invoking, or troubleshooting this skill in an agent host | [agent-hosts.md](references/agent-hosts.md) |
| A copyable template for a pure-server App (single entity + sqlite + actor auth): `.skel` file rules, repo/impl/app, testkit, and seed | [example-greeting](references/example-greeting/README.md) |

## Map the Complete Change Path

Before writing code, map the request to the actual affected points in this path:

```text
Skel contract → generated code → App capability wiring → execution/DI → Link → Hub/Portal/NATS → validation
```

Mark only the nodes that are actually affected. Do not mechanically run every validation just because a node appears in the path. Before settling on a boundary, confirm the consumers, failure semantics, identity, timeouts, idempotency, durability, and the target deployment topology.

## Implement Changes

The following implementation rules come from the Vine `v0.13.1` source and the website's `next` documentation as researched on 2026-08-18. When the target project uses a different version, first verify construction timing, hook/panic behavior, configuration, retries, locks, and transport behavior at the pinned revision. When they differ, defer to the target revision's public source, GoDoc, tests, and release notes.

### Use Generated Boundaries

- For tasks covered by the generation gate, complete the checks, generation, and diff review above before writing implementation that depends on the new boundary.
- Use the generated clients, default Servers, Listeners, Runners, emitters, launchers, and configuration types. Do not hand-write parallel DTOs, interfaces, or protocol adapters.
- With Vine `v0.13.1`, skelc `v0.12.0` or later generates typed in-process request/result clone hooks. Older generated specs use Vine's serialization fallback. Preserve generated `MethodSpec` ownership and test value isolation; use a real Rpc transport when validating JSON/CBOR behavior.

Do not import `go.yorun.ai/vine/internal/*` or copy internal transport, registry, or executor implementations. Ordinary applications should use only `app`, `core/*`, `infra/*`, `util/*`, and the generated facades.

### Organize Business Services by Entity

- First identify the business entities, aggregates, and standalone use cases inside the App, then declare a concrete name for each boundary (e.g. a CRM's `CustomerService`, `DealService`, `ActivityService`, and `DashboardService`). Do not create an App-named umbrella Service.
- Give each Service an independent, same-named, traceable contract file in `skel/`, and provide the corresponding adapter, business, and repository implementations under `src/server/impl/`, `src/server/core/`, and `src/server/repo/` for the same entity or use case.
- Use explicitly named orchestration Services for cross-entity flows; use use-case names such as `DashboardService` or `ReportService` for read-only summaries.
- `ServicerInitHandlers` registers multiple concrete Handlers of the same App. Sharing a database Component does not justify merging Repository structs; per-entity Repositories may reuse the same connection but keep independent interfaces and responsibilities.

### Decide on the Frontend After the Server

- Complete the server contract, generation, implementation, tests, and layout validation first. If the original request did not mention a frontend, ask the user after the server is complete. Do not create `src/web/` or frontend configuration before receiving an explicit answer.
- Once the user confirms a frontend, create the browser package in `src/web/`, leave `src/server/` untouched, and read [frontend-vrpc.md](references/frontend-vrpc.md) for the generated TypeScript vRPC, Portal WEBGW, state layer, and validation. Only now copy and adapt the [Linux launcher](scripts/start_vine_app.sh) and the [Windows launcher](scripts/start_vine_app.bat) (replace the template's `demo` with the real project name; set `VINE_PREPARE_PACKAGE=./cmd/migrate` only when the project actually contains that migration entry).

### Wire the Application

- Treat `src/server/` as the server Go module root, keep `go.mod` and `go.sum` there, and run Go commands from that directory. Maintain a separate `go.mod` under `skeled/golang/`; require that generated-code module at `v0.0.0` from the server module and use `replace <module>/skeled/golang => ../../skeled/golang`. Do not copy generated Go code into `src/server/`.
- Keep `src/server/cmd/<name>/main.go` thin: it only selects the runtime mode and starts the application. Put the App definition and dependency wiring in `src/server/app/app.go`.
- Make the application name match `^[a-z]+(?:\.[a-z]+)*$`. Declare Components and Modules as pointer types, each type only once.
- Use Components for databases, Redis, and other infrastructure. Use Modules for business services, workers, and business lifecycle resources.
- Keep `BindCommon`, Component `Bind`, and Module `Bind` deterministic and free of side effects. Do not start resources or register routes during binding.
- Decide construction inputs such as listener addresses through constructor arguments or the App spec's `DIInit()`. Modifying them in `BindCommon` is already too late.
- Wire only one `WebberSpec` per App, and append multiple Web Handlers to it.
- A local Standalone should explicitly pass a stable, project-specific `DashboardURL`, and use the same address in the README/startup script. Parse and inspect the listening port before startup. If another process owns the port, stop; do not kill the process or silently fall back to `:7099`.
- Make Standalone load the checked-in Hub configuration from `src/server/seed/hub.yaml`. Do not scatter project configuration into other directories.

### Preserve the Dashboard and Hub Version Boundary

- The Dashboard is the Vine Hub control plane, not the business frontend. Do not treat a working Dashboard as proof that the business UI has been delivered.
- Dashboard static assets, Hub, Portal, and Link binaries, and the Go module must come from the same compatible Vine revision.
- When changing the Vine version, Dashboard assets, Hub/Portal/Link, control-plane schema, Standalone configuration, or control-plane routes, send at least one control-plane request from the real Dashboard origin.
- When changing a public Rpc schema, capability registration, Portal entry/admission, or deployment topology, invoke at least one affected business Rpc through a real Portal/vRPC path.
- Pure App-internal algorithm, DAO, or logging changes may mark Dashboard/Portal validation as not applicable, but must not claim that the complete runtime path passed.
- If a response contains `rpc service schema is not found: X`, compare the `X` the browser actually requested with the full Skel name registered by the target revision. If they differ, fix the version, assets, origin, or cache; do not fabricate an alias service in the business App.
- When upgrading Vine on the same origin, verify both uncached sessions and existing browser sessions, and record the steps to clear site caches or force a reload.

### Preserve Lifecycle and Execution Scope

- Put readiness checks and bounded warmup in `BeforeAppStart`. Requests may already arrive by `AfterAppStart`.
- In `AfterAppStart`, start only background loops that remain safe after the application becomes visible, and retain cancel and join mechanisms.
- In `BeforeAppStop`, stop producers, cancel and wait for workers. Release final resources in `AfterAppStop`.
- Do not assume startup failures roll back automatically. A `BeforeAppStart` error becomes a panic.
- Do not let Modules, Components, goroutines, or caches retain execution-scoped objects obtained by Handlers or filters.
- Continue to use the injected context for downstream calls. Do not replace an active request with `context.Background()`.
- Preserve onion ordering in Filters. A short-circuit must produce results compatible with the target protocol.

### Preserve High-Risk Semantics

- `instant` configuration affects only later DI resolutions; it does not update already-injected pointers in place.
- Event and Task delivery may retry by default. Put a stable business ID in the contract, keep side effects idempotent, and observe cancellation.
- Embedded NATS provisions memory-backed `VINE_EVENTS` and `VINE_TASKS`. External NATS must pre-provision both streams before Hub or Link starts and owns storage/replication. Do not promise Event/Task ordering, DLQs, replay, synchronous results, or restart durability without verified file-backed recovery evidence.
- Vine RDB does not migrate business tables automatically. Treat migrations as an explicit deployment step.
- A Redis lock is an expiring coordination lease without a fencing token. Observe `Lock.Context()`; do not unconditionally `defer Unlock()` for work that may lose the lock.
- Use `core/ex` to express stable cross-boundary errors. Do not construct unregistered error codes.
- Use Skel `@sensitive` and `core/redact` to protect logs. Do not log credentials, private keys, or raw sensitive payloads.
- Keep App-to-Link h2c within a loopback or same-host trust boundary. Do not expose internal runtime listeners to untrusted networks.

## Diagnose Problems

Follow the shortest runtime path and locate the first failing boundary:

1. Do versions and generated code match?
2. Does App construction, naming, DI, and lifecycle succeed?
3. Are the schema and capabilities actually wired and registered?
4. Are the local App endpoint and Link API reachable?
5. Have Link's local state, distributed discovery, and configuration snapshots converged?
6. Do the Hub database, Redis distribution, and NATS satisfy the current capability?
7. Do the Portal entry, site, certificate, and target Link match?
8. Do trace, deadline, Actor, structured errors, and logs survive across boundaries?

Reproduce in the smallest topology that covers the failure. Do not use standalone to prove lease expiry, independent process crashes, cross-host reachability, TLS, or real network draining. When only diagnosis is requested, give the root cause, evidence, and repair recommendation without implementing the fix.

## Validate Changes

Run validation according to impact and retain raw evidence:

| Impact | Minimum validation |
| --- | --- |
| New-project skeleton | Verify root-level `skel/*.skel`, `skeled/golang/`, `skeled/typescript/`, the `src/server/` directories, and `src/server/seed/hub.yaml`. Confirm that `skeled/golang/go.mod` exists and that `src/server/go.mod` requires it at `v0.0.0` with a `replace` to `../../skeled/golang`. Confirm that no root-level server directories, `seed/`, or `web/` exist, and that `src/web/` exists only after the frontend is confirmed |
| Business service organization | Verify that every public Service corresponds to a concrete entity or named use case; the names in `skel/`, `src/server/impl/`, `src/server/core/`, and `src/server/repo/` are traceable, the App registers all concrete Handlers, and no God Service named after the App aggregates every entity's logic |
| Skel config usage | Verify that dependents inject the generated `*skeled.<Name>Config` directly; no field-mirroring `core.Config`, `NewConfig`, or `BindCommon` factory that exists only to convert the Config |
| `.skel` or generated code | Pinned `skelc check`, project generation commands, and generated-diff review |
| Go implementation | `gofmt`, targeted package tests; run `go test ./...` when risk permits |
| Public API, reflection, concurrency, or runtime wiring | Targeted tests, `go test ./...`, `go vet ./...` |
| Rpc/Web/Event/Task/DI | `app/testkit` behavior tests; one test package shares one standalone runtime |
| In-process Rpc or generated clone changes | Standalone value-isolation tests plus a `vine dev` or separated JSON/CBOR wire test when codec behavior matters |
| Vine upgrade, Dashboard assets, Hub/Portal/Link, or control-plane schema | Project-specific Dashboard origin, port ownership check, real control-plane request, and full service Skel name verification |
| External NATS Event/Task | Pre-provision `VINE_EVENTS`/`VINE_TASKS`; verify subjects, retention, storage, replicas, reconnect, restart recovery, and idempotency |
| Public Rpc schema, capability registration, Portal admission, or deployment topology | Real Portal/vRPC business call verifying Actor, trace, status mapping, and target service name |
| Browser frontend delivery | `src/web` typecheck, test, lint, and build; combined launcher `--check`; real Portal WEBGW/vRPC; the final reply includes copyable required versions, first/subsequent/Windows startup commands, the preflight-only command, and access addresses |
| Retry, timeout, or graceful stop | Forced failure, cancellation, duplicate delivery, and stop tests |
| Discovery, leases, TLS, or networking | Linked/separated multi-process tests matching the target topology |
| Deployment | Pinned-version output, staging end-to-end path, rolling rollout, and failure drills |

Finally run `git diff --check`. Confirm that no generated files were hand-edited, no new `internal/*` imports were added, no sensitive configuration leaked, and the user's existing changes were not modified unintentionally. Validation that cannot run must state why; never report an unrun check as passing.

When the goal is to deliver a runnable service, upgrade Vine, change the control plane, or deploy a production topology, missing the corresponding Dashboard/Portal end-to-end validation is blocking. For other tasks, record the incomplete end-to-end validation as residual risk and avoid claiming that the complete service path passed.

## Deliver the Result

Lead with the outcome, then list:

1. The identified Go, Vine, and skelc versions and the runtime topology.
2. Which boundaries the change or root cause crosses.
3. The checks actually run and their results.
4. Any unverified version, network, durability, security, or operational risk.
5. If a browser frontend was delivered, directly give the required versions, Linux/macOS first startup, subsequent startup, Windows startup, preflight-only, the business page/vRPC/Dashboard addresses, and any known startup blockers.
