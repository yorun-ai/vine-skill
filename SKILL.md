---
name: vine-skill
description: Develop, troubleshoot, review, test, upgrade, or deploy Yorun Vine services and browser frontends. Changes to capability contracts, generated boundaries, capability registration, or Vine/skelc versions require pinned Skel checks, generation, and diff review; implementation-only changes must preserve and validate the existing generation baseline. Use Rpc/vRPC for structured synchronous APIs. Browser frontends must use the project vRPC client with skel-generated TypeScript services, specs, and types, and browser delivery must use a generated Web capability through Portal WEBGW; reserve Web/HTTP for binary streams and static assets. Use for projects importing go.yorun.ai/vine or involving ApplicationSpec, Component, Module, DI, lifecycle, generated Rpc/Web/Event/Task/config contracts, TypeScript web clients, React Query, app/standalone, app/linked, Hub, Link, Portal, RPCGW, WEBGW, Vine CLI, Redis, RDB, compatibility, or production readiness.
---

# Develop Vine Applications

## Follow the Core Boundaries

Always follow these principles:

1. **Pin versions first**: Determine the project's pinned Go, Vine, and skelc versions before selecting APIs, CLI flags, or generation commands. Before Vine 1.0, the website's `next` documentation may be ahead of released versions.
2. **Apply the generation gate by impact**: When adding or changing a Skel contract, generated boundary, capability registration, or Vine/skelc version, complete the pinned `skelc check`, code generation, and generated diff review first. For implementation-only changes, do not modify generated files; re-enter the gate immediately if the generation baseline is inconsistent.
3. **Prefer generated vRPC**: Declare structured synchronous business calls as Rpc and invoke them through generated clients or Portal vRPC. Browser frontends must use the project's vRPC client with skel-generated TypeScript services, specs, and types. Use Web/HTTP only for file uploads, downloads, Range, multipart, other binary streams, or static asset delivery.
4. **Separate responsibilities**: Let App declare capabilities, Module own business behavior, Component own infrastructure, Link handle discovery and delivery, Hub manage control-plane state, and Portal admit external traffic.
5. **Keep scopes correct**: Keep request-related contexts, clients, configuration, DAOs, caches, and lockers within the execution that created them.
6. **Be honest about topology**: Treat standalone results only as single-process evidence. Use linked or separated topologies to validate networking, leases, draining, TLS, and independent failures.
7. **Keep the control plane on one revision**: Treat Dashboard assets, Hub schema, and Portal routes as one versioned unit. Do not reuse a Dashboard origin across projects or Vine revisions during local development.
8. **Act only as authorized by the request**: Stay read-only for review or diagnosis. Modify code and perform the corresponding validation only when implementation is explicitly requested.
9. **Deliver local browser apps through Portal**: For a new browser-enabled Standalone project without an established topology, make `http://127.0.0.1:7288/` the public origin: route `/api` to RPCGW and `/` to the exact generated Web capability through WEBGW. Keep Dashboard on a separate project-specific origin such as `:7299` and the Vite development upstream on `:5174`. Do not add this topology to backend-only work.
10. **Provide native startup scripts**: A delivered browser-enabled local project must include `scripts/start_vine_app.sh` for Linux and `scripts/start_vine_app.bat` for Windows. Do not require Python merely to start the service.

## Decide Whether the Generation Gate Applies

Before implementing, determine whether the change touches a contract or generated boundary.

Enter the generation gate for any of the following:

- Add, remove, or modify `.skel` data types, Actors, errors, Rpc, Web, Event, Task, or config.
- Change public methods, parameters, return values, full Skel names, serialization shapes, or cross-boundary errors.
- Add, remove, or adjust App capability registration so that a schema or capability becomes exposed or withdrawn.
- Change Vine, skelc, generation commands, generation parameters, generated directories, or generated package boundaries.
- Encounter generated code that is missing, stale, uncompilable, or inconsistent with the pinned contract.

After entering the gate, follow this exact order:

1. Confirm the project's pinned Go, Vine, and skelc versions and its existing generation command.
2. Modify hand-maintained `.skel` files first; leave them unchanged when the contract does not need to change.
3. Run `check` with the pinned skelc version.
4. Regenerate code with the project's existing command.
5. Review the generated diff, schema, Go Server/Client, TypeScript service/spec/types, Listener, Runner, and all affected package boundaries.
6. Only after the first five steps succeed, modify Handlers, capability wiring, call adapters, and other implementations that depend on the generated boundary.
7. If the contract changes again, return to steps 2–5.

Only tasks covered by this gate are blocked when skelc is unavailable, versions do not match, checks fail, or generation fails. Never fabricate generated types, hand-write parallel DTOs, interfaces, or routes, or promise to add Skel later.

An implementation-only change must satisfy all of these conditions:

- It does not change `.skel`, generated code, the public wire shape, capability registration, Vine/skelc versions, or generation commands.
- It only changes business algorithms behind an existing contract, DAO queries, logging, internal services, lifecycle behavior, or other hand-written behavior.
- It leaves the generated directory unchanged and confirms the existing generation baseline through `go.mod`, generated-file markers, and the project build.
- It runs tests that match the impact. If the schema, generated types, or registration are inconsistent, stop and re-enter the generation gate.

## Select the Capability Boundary

| Need | Choice |
| --- | --- |
| Collaboration inside one App without a capability boundary | Plain Go interfaces and DI |
| Structured synchronous business requests/responses across Apps, browsers, or external clients | Skel Rpc with generated clients; use vRPC at external entries |
| Multipart, upload, download, Range, `Content-Disposition`, binary streams, or static assets | Web/HTTP |
| Broadcast a fact that has already occurred | Event |
| Ask one available worker to perform work | Task |
| Declarative runtime parameters | Skel config |
| Caching or distributed coordination | Redis Component |
| Relational persistence | RDB Component with separately managed schema migrations |

Do not create JSON REST Web routes for CRUD, queries, search, pagination, authentication, or state changes. When one feature includes both metadata and file bytes, use Rpc/vRPC for metadata, authorization, and the transfer session, then Web/HTTP for the bytes, bound to the same identity and business ID. Choose Rpc when uncertain.

## Establish the Project Baseline First

Before modifying anything:

1. Read `AGENTS.md`, `README`, and contribution instructions in the project root and relevant subdirectories.
2. Inspect the worktree and protect the user's existing changes.
3. Locate every `go.mod`, `.skel` file, generated directory, App specification, process entry point, and deployment configuration.
   For frontend work, also locate the relevant `package.json` files, generated TypeScript package and exports, shared vRPC client, Portal origin/configuration, and existing query layer.
4. Confirm actual versions in the target Go module:

```bash
go version
go list -m -json go.yorun.ai/vine
vine version --json
skelc version
```

Run only commands already available locally. Do not install or upgrade tools merely to detect versions. If commands are unavailable, gather evidence from `go.mod`, `go.sum`, CI, generated files, and build scripts, and state what remains unknown.

5. Identify the current runtime mode:
   - `app/standalone`: embeds Hub, Portal, and Link.
   - `app/linked`: runs the business App and Link in one process and connects to an external Hub.
   - `app.New`: connects the business App to a separate Link.
   - `vine dev`: runs a local runtime process connected to a separate business App.
6. Report the version and topology baseline. Do not treat examples from the current website as the project's API automatically.
7. When using the Standalone Dashboard, check whether another process owns the target host/port and explicitly set a project-specific `DashboardURL` in code or startup configuration. Do not rely on the default `:7099` or reuse a browser origin that has served another Vine revision.
8. For a browser-enabled local Standalone app, locate both Portal sites and rules. Unless the project already pins another topology, reserve `:7288` for the public business origin, use `/api` for RPCGW and `/` for WEBGW, keep Dashboard separate, and stop with an ownership report if a required port is occupied. Do not kill the owner or silently change ports.

See [foundations.md](references/foundations.md) for detailed baseline checks.

## Load References by Problem

Read only the references needed for the current task. Combine them for cross-domain changes.

| Problem | Read |
| --- | --- |
| Versions, first app, Skel workflow, project structure, App, Component, Module, configuration | [foundations.md](references/foundations.md) |
| Rpc, Web, Event, Task, and vRPC HTTP | [capabilities.md](references/capabilities.md) |
| Browser frontend, generated TypeScript vRPC clients, React Query, and Portal validation | [frontend-vrpc.md](references/frontend-vrpc.md) |
| DI, execution, Filter, Meta, Actor, trace, timeout, and errors | [execution-and-boundaries.md](references/execution-and-boundaries.md) |
| Redis, RDB, logging, redaction, and testkit | [data-and-testing.md](references/data-and-testing.md) |
| Lifecycle, routing, topology, Hub, Link, Portal, CLI, security, rollout, and failure drills | [runtime-operations.md](references/runtime-operations.md) |
| Exact website pages, source, release notes, or refreshed version facts | [official-sources.md](references/official-sources.md) |

## Map the Complete Change Path

Before writing code, map the request to the actual affected points in this path:

```text
Skel contract → generated code → App capability wiring → execution/DI → Link → Hub/Portal/NATS → validation
```

For a browser feature, also map the concrete call path:

```text
Skel Rpc → generated Go Server + generated TypeScript service/spec/types → project vRPC client → Portal → Link → App → frontend state/UI

Skel Web → generated Default...WebServer → app.WebberEnabled → WebberInitHandlers
         → WEBGW site/rule → Portal → Link → App Web handler → reverse proxy or embedded assets
```

Mark only the nodes that are actually affected; do not mechanically run every validation merely because a node appears in the path. Before selecting the boundary, confirm consumers, failure semantics, identity, timeouts, idempotency, durability, and the target deployment topology.

## Implement Changes

The following implementation rules come from the Vine `v0.12.0` source and the website's `next` documentation as researched on 2026-08-03. If the target project uses another version, verify construction timing, hook and panic behavior, configuration, retries, locks, and transport behavior at the pinned revision. Prefer public source, GoDoc, tests, and release notes from the target revision when they differ.

### Use Generated Boundaries

- For tasks covered by the generation gate, complete the checks, generation, and diff review above before writing code that depends on the new boundary.
- Use generated clients, default Servers, Listeners, Runners, emitters, launchers, and configuration types. Do not hand-write parallel DTOs, interfaces, or protocol adapters.
- For implementation-only changes, leave `.skel` and generated directories unchanged. Stop and re-enter the gate if the generation baseline is abnormal.

Do not import `go.yorun.ai/vine/internal/*` or copy internal transport, registry, or executor implementations. Ordinary applications should use only `app`, `core/*`, `infra/*`, `util/*`, and generated facades.

### Build Browser Frontends Through Generated TypeScript vRPC

- Read [frontend-vrpc.md](references/frontend-vrpc.md) before implementing or reviewing a browser feature that calls a Vine service.
- Import the actual service factory, `ServiceSpec`, request/result types, and errors from the target project's skel-generated TypeScript package. Treat its checked-in generated output as authoritative and never edit it by hand.
- Pass the project's shared vRPC client to the generated service factory. Do not call App or Link internal endpoints, recreate service/method names, or implement structured business calls with hand-written `fetch`, Axios, REST routes, or duplicate DTOs.
- Keep server state in the project's existing query layer. When React Query is used, derive query and mutation key prefixes from generated `ServiceSpec.serviceName` and `ServiceSpec.methods`; pages consume hooks rather than assembling RPC requests.
- If the contract changes, complete pinned `skelc check`, regenerate both affected Go and TypeScript boundaries, and review the diff before updating the frontend. If only UI composition changes, leave `.skel` and generated files untouched.
- A browser UI also requires a generated Skel `web` capability for asset delivery. Complete the full path: Skel Web → pinned `skelc check`/generation/diff → generated `Default...WebServer` → `app.WebberEnabled` → `WebberInitHandlers` → WEBGW site/rule → Portal/Link/App registration. Do not put a hand-written HTTP server beside Vine.
- For new local Standalone browser apps, use one public origin: `http://127.0.0.1:7288/` for the page and `/api/invoke` for vRPC. The `/api` RPCGW rule and `/` WEBGW rule share port `7288`; Vine's longest matching path keeps API calls out of the frontend catch-all.
- During development, own `web.NewReverseProxy` in a Module lifecycle, register `ANY("/*path")` on the generated Web handler, default the upstream to `http://127.0.0.1:5174`, close the proxy during App shutdown, and return `503 Service Unavailable` with `Retry-After: 1` when Vite is unavailable. In production, build the frontend and serve its output with `web.NewAssetsServer` instead of running a Vite development server.
- Copy and adapt the bundled [Linux starter](scripts/start_vine_app.sh) and [Windows starter](scripts/start_vine_app.bat) into the target project's `scripts/` directory. Both must check required files and commands, refuse occupied `5174`/`7288`/`7299`, optionally run `pnpm install` with `--install`, start Vite before Vine, wait for the public page, and stop both child process trees. Keep their environment overrides and behavior aligned. Never infer or automatically run a business migration; a project may explicitly set `VINE_PREPARE_PACKAGE=./cmd/migrate`.

### Wire the Application

- Keep `cmd/<name>/main.go` thin; it should only select the runtime mode and start the application.
- Make the application name match `^[a-z]+(?:\.[a-z]+)*$`.
- Declare Components and Modules as pointer types, and declare each type only once.
- Use Components for databases, Redis, and other infrastructure. Use Modules for business services, workers, and business lifecycle resources.
- Keep `BindCommon`, Component `Bind`, and Module `Bind` deterministic and free of side effects. Do not start resources or register routes while binding.
- Set listener addresses and other construction inputs through constructor arguments or the App spec's `DIInit()`. Changing them in `BindCommon` is too late.
- Wire only one `WebberSpec` per App and append multiple Web Handlers to it.
- For local Standalone, explicitly supply a stable, project-specific `DashboardURL` and use the same address in the README or startup script. Parse and inspect the listener port before startup. If another process owns it, stop; do not kill that process or silently fall back to `:7099`.
- When the App includes a browser frontend, make `standalone.Option.SeedYAMLFile` load the checked-in RPCGW/WEBGW seed. The WEBGW site's `webName` must exactly match the generated full Skel Web name. Keep the Dashboard off the public `:7288` business origin, for example on `http://127.0.0.1:7299/`.

### Preserve the Dashboard and Hub Version Boundary

- The Dashboard is the Vine Hub control plane, not the business frontend. Do not treat a working Dashboard as proof that the business UI has been delivered.
- Dashboard assets, Hub, Portal, and Link binaries, and the Go module must come from a compatible Vine revision.
- When changing the Vine version, Dashboard assets, Hub/Portal/Link, control-plane schema, Standalone configuration, or control-plane routes, send at least one control-plane request from the real Dashboard origin.
- When changing a public Rpc schema, capability registration, Portal entry/admission, or deployment topology, invoke at least one affected business Rpc through the real Portal/vRPC path.
- For changes confined to App-internal algorithms, DAOs, or logging, mark Dashboard/Portal validation as not applicable, but do not claim that the complete runtime path passed.
- If a response contains `rpc service schema is not found: X`, compare the browser's requested `X` with the full Skel name registered by the target revision. If they differ, fix the version, assets, origin, or cache; do not fabricate alias services in the business App.
- When upgrading Vine on the same origin, verify both uncached and existing browser sessions, and record steps to clear site data or force a reload.

### Preserve Lifecycle and Execution Scope

- Put readiness checks and bounded warmup in `BeforeAppStart`. Requests may already arrive during `AfterAppStart`.
- In `AfterAppStart`, start only background loops that remain safe after the application becomes visible, and retain cancellation and join mechanisms.
- In `BeforeAppStop`, stop producers, cancel workers, and wait for them. Release final resources in `AfterAppStop`.
- Do not assume startup failures roll back automatically. A `BeforeAppStart` error becomes a panic.
- Do not let Modules, Components, goroutines, or caches retain execution-scoped objects obtained by Handlers or filters.
- Continue using the injected context for downstream calls. Do not replace an active request with `context.Background()`.
- Preserve onion ordering in Filters. A short-circuit must produce results compatible with the target protocol.

### Preserve High-Risk Semantics

- `instant` configuration affects later DI resolutions; it does not mutate an already injected pointer.
- Event and Task delivery may retry by default. Put a stable business ID in the contract, make side effects idempotent, and observe cancellation.
- Do not promise unverified Event/Task ordering, DLQs, replay, synchronous results, or disk durability.
- Vine RDB does not migrate business tables automatically. Treat migrations as an explicit deployment step.
- A Redis lock is an expiring coordination lease without a fencing token. Observe `Lock.Context()` and do not unconditionally `defer Unlock()` for work that may lose the lock.
- Use `core/ex` for stable cross-boundary errors. Do not invent unregistered error codes.
- Use Skel `@sensitive` and `core/redact` to protect logs. Do not log credentials, private keys, or raw sensitive payloads.
- Keep App-to-Link h2c inside a loopback or same-host trust boundary. Do not expose internal runtime listeners to untrusted networks.

## Diagnose Problems

Follow the shortest runtime path and locate the first failing boundary:

1. Do versions and generated code match?
2. Do App construction, naming, DI, and lifecycle succeed?
3. Are the schema and capabilities actually wired and registered?
4. Are the local App endpoint and Link API reachable?
5. Have Link local state, distributed discovery, and configuration snapshots converged?
6. Do the Hub database, Redis distribution, and NATS satisfy the current capability?
7. Do the Portal entry, site, certificate, and target Link match?
8. Do trace, deadline, Actor, structured errors, and logs survive each boundary?

For `webgw endpoint is unavailable: <site>`, do not debug Vite first. This message means Portal selected a WEBGW site but the exact configured `webName` is not registered by an App on the selected Link. Check the `.skel` Web declaration and generated output, the full Skel name, `app.WebberEnabled`, `WebberInitHandlers`, Link/App registration convergence, and whether an old process is still running. A registered Web handler with an unavailable Vite upstream should instead return the explicit `503` fallback.

Reproduce in the smallest topology that covers the failure. Do not use standalone to prove lease expiry, independent process crashes, cross-host reachability, TLS, or real network draining. When the request is diagnosis only, provide the root cause, evidence, and repair recommendation without implementing the fix.

## Validate Changes

Run validation according to impact and retain raw evidence:

| Impact | Minimum validation |
| --- | --- |
| `.skel` or generated code | Pinned `skelc check`, project generation command, and generated diff review |
| Go implementation | `gofmt`, target package tests, and `go test ./...` when risk permits |
| Public API, reflection, concurrency, or runtime wiring | Targeted tests, `go test ./...`, and `go vet ./...` |
| Rpc/Web/Event/Task/DI | `app/testkit` behavior tests; share one standalone runtime per test package |
| Browser frontend calling a Vine service | Frontend typecheck/build and targeted tests; `bash -n` plus Linux starter preflight; Windows Batch review and a Windows CI/host preflight when available; Web handler proxy/assets tests; a dynamic-port Standalone test using the real seed and `Portal → WEBGW → Link → App Web → test frontend`; plus a real call through the generated TypeScript service and Portal/vRPC entry. Verify `/api` wins over `/`, service/method and Web identities, Actor policy, status/error mapping, query invalidation, and the `503` upstream fallback |
| Vine upgrade, Dashboard assets, Hub/Portal/Link, or control-plane schema | Project-specific Dashboard origin, port ownership check, real control-plane request, and full service Skel name verification |
| Public Rpc schema, capability registration, Portal admission, or deployment topology | Real Portal/vRPC business call verifying Actor, trace, status mapping, and target service name |
| Retry, timeout, or graceful stop | Forced failure, cancellation, duplicate delivery, and stop tests |
| Discovery, leases, TLS, or networking | Linked/separated multi-process tests matching the target topology |
| Deployment | Pinned version output, staging end-to-end path, rolling rollout, and failure drills |

Finally run `git diff --check`. Confirm that generated files were not hand-edited, no new `internal/*` imports were added, sensitive configuration was not leaked, and existing user changes were not modified unintentionally. Review every Web route and frontend transport call: treat any structured JSON business request/response, hand-written Rpc path, or duplicated generated DTO as a boundary error and replace it with the skel-generated TypeScript vRPC client. State why any validation could not run; never report an unrun check as passing.

When the goal is to deliver a runnable service, upgrade Vine, change the control plane, or deploy a production topology, missing the corresponding Dashboard/Portal end-to-end validation is blocking. For other tasks, record missing end-to-end validation as residual risk and do not claim that the complete service path passed.

## Deliver the Result

Lead with the outcome, then list:

1. The identified Go, Vine, and skelc versions and runtime topology.
2. The boundaries crossed by the change or root cause.
3. The checks actually run and their results.
4. Any unverified version, network, durability, security, or operational risk.
