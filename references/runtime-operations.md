# Vine Runtime, Deployment, and Production Boundaries

This reference is based on Vine `v0.13.1`. Use it for topology, routing, readiness, Hub/Link/Portal, CLI, mTLS, durability, rollout, and failure-drill tasks. CLI flags, default ports, and time constants may change by version. Verify them first with the target binary's supported `vine version` and `vine <command> --help`; do not assume that `version` accepts `--json`. The liveness timings, ports, and drain durations here are `v0.13.1` version facts; the authoritative list is in [Version baseline quick reference](foundations.md#version-baseline-quick-reference).

## Contents

- [Roles and state flow](#roles-and-state-flow)
- [Select a topology](#select-a-topology)
- [Registration, routing, and readiness](#registration-routing-and-readiness)
- [Lifecycle and draining](#lifecycle-and-draining)
- [CLI baseline](#cli-baseline)
- [Local browser business origin](#local-browser-business-origin)
- [Security boundaries](#security-boundaries)
- [Durability and asynchronous delivery](#durability-and-asynchronous-delivery)
- [Production validation checklist](#production-validation-checklist)

## Roles and State Flow

> Hub knows the state; Link connects applications; Portal admits external traffic; App executes business code.

| Role | Owns | Synchronous business path |
| --- | --- | --- |
| App | Components, Modules, Handlers, Listeners, Runners, and business state | Caller or target |
| Link | Local App source state, configuration, discovery, forwarding, Event/Task consumers, and draining | Yes |
| Hub | Configuration, schemas, registrations, leases, Portal configuration, and runtime distribution | No |
| Portal | External listeners, sites, TLS, admission, and endpoint selection | External traffic only |
| NATS | Event/Task messages | Asynchronous only |

Key flows:

```text
Config: seed/operator → Hub DB → Redis snapshot/change → Link reader → execution DI

Internal Rpc: caller App → caller Link → local App or target Link ingress → target App

External Rpc/Web: client → Portal entry/site/admission → selected Link ingress → App

Async: emitter/launcher → sender Link → NATS → consumer Link → Listener/Runner App
```

The Hub database is the source of truth for managed configuration and Portal rules, sites, and certificates. Redis distributes runtime snapshots and changes; it is not a database replacement. App registration contains identity, endpoints, schemas, and declared capabilities, not business data.

## Select a Topology

| Mode | Process boundary | Suitable for | Does not cover |
| --- | --- | --- | --- |
| Standalone | Hub, Portal, Link, and App in one process | Learning, one App, local integration tests | Heartbeats, TTL, independent failures |
| `vine dev` | Runtime in one process, business App separate | A real local App-to-Link network boundary | Hub leases and independent runtime failures |
| Linked | External Hub/Portal, Link and App in one process | Shared control plane with fewer sidecars to operate | Independent App and Link lifecycles |
| Separated | Hub, Portal, Link, and App in separate processes | Container deployment, real failures, and scaling | Automatic HA or security configuration |

Start with standalone. Move to linked when runtime state must be shared, and to separated when real process isolation, failures, and scaling are required. Each business App and its Link sidecar should run on the same host or Pod, share a trust boundary, and scale together.

Only the constructor changes across process modes:

```go
// Standalone
standalone.New[*checkout.App]().StartAndWait()

// Separated App
app.New[*checkout.App]().StartAndWait()
```

Business packages, generated contracts, Components/Modules, and Handlers should not be rewritten for a topology change.

## Registration, Routing, and Readiness

Registration path:

1. App constructs Components, Modules, and capability servers.
2. App runs every `BeforeAppStart` hook.
3. App starts endpoints and capability mechanisms.
4. App registers identity, endpoints, schemas, and Rpc/Web/Event/Task capabilities with Link.
5. Link installs local state and publishes it to Hub.
6. Remote Links and Portal converge asynchronously on the snapshot.
7. App runs `AfterAppStart`.

Therefore:

- Requests may arrive during `AfterAppStart`.
- `Start()` returning and a local listener being open do not mean remote or Portal routing has converged globally.
- Use the real caller path as the readiness probe.
- An App with no public capabilities publishes no capability registration, while local lifecycle still runs normally.

Routing selection:

- Link and Portal maintain round-robin cursors over their own snapshots.
- There is no global ordering, weight, affinity, circuit breaker, real-time pre-call probe, or automatic same-request failover.
- Failure of the selected endpoint returns from the current call. Retry explicitly only when idempotency and deadline allow it.
- For Web, Portal selects a target Link from distributed endpoints, while that Link's webproxy indexes only local Apps.

Network-mode liveness facts in `v0.13.1`:

| Behavior | Current value |
| --- | --- |
| Link heartbeat | 10 seconds |
| Hub lease | 30 seconds |
| Hub sweep | 5 seconds |
| Separated Link App health interval | 5 seconds |
| Console ping timeout | 2 seconds |
| Unregistration threshold | 3 consecutive non-timeout failures |

Invocation timeout only logs and does not increment health failures; a stuck App may still have its lease renewed by Link. These are `v0.13.1` implementation constants (see [Version baseline quick reference](foundations.md#version-baseline-quick-reference) for the full list), not CLI tuning contracts.

Standalone disables TTL, heartbeat, sweeper, and App health checks. Linked retains Hub leases and heartbeats but skips App health checks because App and Link share a process.

## Lifecycle and Draining

Graceful App shutdown:

1. Run Module `BeforeAppStop` hooks in reverse order.
2. Run Component `BeforeAppStop` hooks in reverse order.
3. Unregister through Link and drain.
4. Gracefully stop HTTP or remove the in-process route.
5. Cancel the App root context.
6. Run Module and Component `AfterAppStop` hooks in reverse order.

Components needed by in-flight work must remain available until draining completes. `BeforeAppStop` stops producers and cancels and joins workers. `AfterAppStop` releases final resources. Hooks have no automatic deadline, and a panic interrupts the remaining shutdown sequence. Canceling `RunFlag.Context` does not automatically call `StopGracefully`.

Wrapper order:

- Linked: start Link → Apps; stop Apps in reverse order → Link.
- Standalone: start Hub → Portal → Link → Apps; stop Apps in reverse order → Link → Portal → Hub.
- Bundle: start in declaration order and stop in reverse order.

Operational shutdown order for separated mode: stop new external traffic → Apps → corresponding Links → Portal → Hub.

Drain and stop timings in `v0.13.1` are implementation details (see [Version baseline quick reference](foundations.md#version-baseline-quick-reference)): Link propagation grace is about 400 ms, in-flight Rpc/Event/Task drains for at most 30 seconds, App unregister RPC waits at most 1 minute, and App HTTP shutdown is about 10 seconds. Do not treat these values as stable CLI contracts. Orchestrator termination grace must cover the target version's actual boundaries.

## CLI Baseline

Check first:

```bash
vine version
vine --help
vine hub --help
vine link --help
vine portal --help
```

Local `v0.13.1` example:

```bash
vine dev \
  --db-sqlite-file ./hub-dev.sqlite \
  --seed-yaml-file ./src/server/seed/hub.yaml

vine hub serve \
  --mq-embedded-nats \
  --db-sqlite-file ./hub.sqlite

vine portal serve \
  --hub-endpoint http://127.0.0.1:7071

vine link serve \
  --api-listen 127.0.0.1:7079 \
  --ingress-listen 127.0.0.1:7082 \
  --hub-endpoint http://127.0.0.1:7071

VINE_LINK_ENDPOINT=http://127.0.0.1:7079 ./business-app
```

Hub must select exactly one database mode, SQLite or PostgreSQL, and exactly one NATS mode, embedded or external. Provide explicit durability for ordinary Hub/standalone. Only `vine dev` without a database uses and cleans up temporary SQLite.

Important current listeners:

| Boundary | `v0.13.1` default/common value | Caller |
| --- | --- | --- |
| Hub Control | `127.0.0.1:7071` | Link, Portal |
| Hub Redis | `127.0.0.1:7072` | Link, Portal |
| Hub Admin/Dashboard | `127.0.0.1:7075` | Control plane, Portal |
| Link App API | `127.0.0.1:7079` | Same-host business App |
| Link ingress | Dynamic; pin it in production | Portal, remote Links, tools |
| App HTTP | `127.0.0.1:0` | Link sidecar |
| Portal Dashboard entry | `:7099` | External control-plane clients |

Vine v0.12 separates Hub Control from Admin and removes the old Hub `--api-listen`. Do not copy removed flags from older deployment configurations. Link still uses its own `--api-listen`.

### Local Dashboard Version Isolation

The default `:7099` is suitable for one-off examples, not for multiple projects or Vine revisions used in turn. Browsers cache HTML and JavaScript by origin. When the same origin serves different Dashboard versions over time, an old page may ask a new Hub for a renamed full service Skel name. A typical symptom is:

```text
rpc service schema is not found: vine.hub.AppConfigService
```

The target revision may actually register `vine.hub.admin.AppConfigService`. This is a Dashboard/Hub version mismatch, not a missing AppConfig Handler in the business service.

When creating a local Standalone service:

1. Derive a dedicated Dashboard host/port stably from the project name and explicitly set `standalone.Option.DashboardURL`. Do not fall back to the default `:7099`.
2. Before startup, use a read-only port check to confirm that no other process owns the listener. If occupied, stop and report the owning process; do not terminate it without authorization.
3. Pin the Go module, CLI/runtime, and Dashboard assets to the same Vine revision. Pin the revision of external Dashboard assets as well.
4. After startup, fetch the page from the real Dashboard origin and execute one control-plane Rpc, then execute one business Rpc through Portal.
5. Record results for both an uncached/new browser session and an existing session. If an old session requests a retired service name after upgrade, move to the project-specific origin and require clearing site data for the old origin.

During diagnosis, use the full service name in the Portal error to look up the generated schema at the target revision. If the frontend's requested name differs from the Hub registration, stop changing business code and fix the version selection, asset source, origin, or cache. Do not add aliases for old Hub control-plane services in the business App; doing so hides incompatibility and expands an unauthorized control plane.

## Local Browser Business Origin

The Dashboard is not the business frontend. For a new browser-enabled Standalone project without an existing port convention, keep these roles separate:

```text
http://127.0.0.1:7288/            → Portal WEBGW → generated App Web → Vite/assets
http://127.0.0.1:7288/api/invoke  → Portal RPCGW → business Rpc
http://127.0.0.1:7299/            → project-specific Dashboard example
http://127.0.0.1:5174/            → private Vite development upstream
```

Configure `/api` and `/` as separate Portal rules on `7288`. Longest path-prefix matching sends `/api/invoke` to RPCGW before the `/` WEBGW catch-all. Load the checked-in seed with `standalone.Option.SeedYAMLFile`; the WEBGW site's `webName` must exactly equal the generated full Skel Web name. The App must embed `app.WebberEnabled` and register the generated concrete Web handler with `WebberInitHandlers`. See [frontend-vrpc.md](frontend-vrpc.md#publish-a-local-standalone-browser-app-on-port-7288) for the full Skel, App, handler, proxy, seed, startup, and production-assets pattern.

Before startup, inspect both the public Portal port and Dashboard port. If either is occupied, stop and report the owner without killing it or silently selecting a replacement. Start Vite and the Vine process separately, then probe the page through `:7288/` and a generated TypeScript vRPC call through `:7288/api/invoke`.

Deliver [start_vine_app.sh](../scripts/start_vine_app.sh) and [start_vine_app.bat](../scripts/start_vine_app.bat) with browser-enabled projects so Linux and Windows users can perform the same checks and coordinated startup without Python. The scripts may optionally run `pnpm install`, but they never install Go/Vine/skelc or infer business migrations.

If Portal returns `webgw endpoint is unavailable: <site>`, the Portal rule is already matching; the exact Web capability is missing from Link/App registration. Verify generation, the full `webName`, `app.WebberEnabled`, `WebberInitHandlers`, convergence, and the running process. An unavailable Vite upstream is a different failure and should produce the App handler's explicit `503` response.

## Security Boundaries

Backend mTLS is optional, but all three flags are required together when enabled:

```text
--mtls-ca-file
--mtls-cert-file
--mtls-key-file
```

Exact X.509-SVID URI SANs:

```text
spiffe://<trust-domain>/vine/daemon/vine.hub
spiffe://<trust-domain>/vine/daemon/vine.link
spiffe://<trust-domain>/vine/daemon/vine.portal
```

Certificates must use the same trust domain, one exact URI identity, and permit client and server authentication. The target version uses TLS 1.3 and exact SPIFFE authorization. With mTLS enabled, reject discovered plaintext endpoints and do not downgrade.

mTLS covers Hub Control/Admin, embedded Redis/NATS, Link ingress, and component proxy clients. It intentionally does not cover the App ↔ Link API, which is unauthenticated h2c inside a same-host sidecar boundary. Keep it on loopback. A special cross-host deployment must add authentication, encryption, and network restrictions at the platform layer.

Other rules:

- External traffic must pass through Portal. Direct App or Link endpoints bypass admission.
- Keep Portal public TLS certificates separate from daemon mTLS identities.
- A temporary self-signed Portal certificate provides bootstrap encryption only; it is not production trust.
- Access to Hub Redis is equivalent to access to runtime configuration and TLS private keys. Do not expose it as a general-purpose Redis service.
- External PostgreSQL and NATS are responsible for their own authentication, TLS, durability, and backups.
- No untrusted network should reach any internal runtime listener.

## Durability and Asynchronous Delivery

- The Hub database is the source of truth for configuration and Portal rules, sites, and certificates.
- Seed YAML imports initial state only; it is not a continuous backup.
- Redis distributes runtime snapshots and changes; it is not the source-of-truth database.
- Embedded NATS provisions `VINE_EVENTS` for `event.>` with interest retention and `VINE_TASKS` for `task.>` with work-queue retention; both use memory storage.
- External NATS requires JetStream and must pre-provision both streams before Hub or Link starts. Vine clients read the existing streams and do not create them or choose stream/consumer storage. Authentication, TLS, storage, replication, and backups belong to the external deployment.
- For restart durability, provision file-backed external streams and test recovery with the production topology. A single-replica example is:

```bash
nats --server "$VINE_MQ_EXTERNAL_NATS_URL" stream add VINE_EVENTS \
  --subjects "event.>" --retention interest \
  --storage file --replicas 1 --defaults

nats --server "$VINE_MQ_EXTERNAL_NATS_URL" stream add VINE_TASKS \
  --subjects "task.>" --retention workqueue \
  --storage file --replicas 1 --defaults
```

- Set replicas to the required redundancy in a cluster; do not copy the single-replica example unchanged.
- Listener/Runner delivery is at least once, may retry forever, and promises no DLQ, replay, or ordering. Business operations must be idempotent and retain required durable records.
- The current documentation defines only a single-Hub control plane, with no active-active or failover protocol. Starting multiple Hubs does not create HA.

Back up the Hub database and restore it in an isolated environment. After restoration, verify configuration, Portal rules, sites, certificates, and subscriptions.

## Production Validation Checklist

1. Pin Go, Vine, skelc, generated code, and the CLI image. Save the raw `vine version` output.
2. Diagram process and network paths for Hub, Portal, every Link/App, PostgreSQL, and Redis/NATS.
3. In production-equivalent separated staging, start Hub → Portal/Links → Apps.
4. Validate readiness through real Portal/Link paths, not just ports or processes.
5. Cover every public entry and at least one App-to-App Rpc.
6. Execute a control-plane Rpc from the Dashboard origin. Confirm that the frontend's full service Skel name exists in the same-revision Hub schema, and test an uncached session when reusing an origin.
7. Validate trace, Actor, Initiator, and remaining deadlines.
8. Validate new executions for `instant` configuration and restart behavior for `eternal` configuration.
9. For external NATS, verify both required streams, retention, storage, and replicas before Hub or Link starts; then force Event/Task errors, timeouts, duplicates, NATS reconnects, and restart recovery and check idempotency.
10. Drill graceful App replacement, abrupt App loss, Link loss and lease expiry, and Link-to-Hub reconnect.
11. Validate SPIFFE identities, certificate rotation/SNI, pinned ingress ports, and the firewall caller set.
12. Back up and restore the Hub database, then inspect all control-plane state.
13. Stop external traffic → Apps → Links → Portal → Hub, and confirm that termination grace is sufficient.

Vine does not provide a universal `/healthz`. In addition to process checks, keep an end-to-end business probe that covers the real Hub/Link/Portal/discovery/forwarding path.

See [official-sources.md](official-sources.md) for official pages and the version refresh path.
