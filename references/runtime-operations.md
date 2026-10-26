# Vine Runtime, Deployment, and Production Boundaries

This reference describes Vine `v0.32.5`. Inspect the pinned CLI and source when using another version. See [baseline-upgrade.md](baseline-upgrade.md) for removed APIs, seed vocabulary and persistent-database migration floors.

## Roles and State Flow

| Role | Owns |
| --- | --- |
| App | Components, Modules, capabilities and business state |
| Link | App registration, local routing, discovery, dispatch and draining |
| Hub | Seed/database configuration, descriptors, leases, Watch distribution and runtime services |
| Portal | Public HTTP/HTTPS entries, site routing, admission and forwarding |
| NATS | Event and Task streams |

```text
Configuration: seed/database → Hub Watch → Link/Portal snapshots
Internal Rpc: caller App → caller Link → local App or target Link → target App
External API/Web: client → Portal entry/site/admission → Link → App
Async: emitter/launcher → Link → NATS → consumer Link → Listener/Runner
```

Hub Watch distributes configuration and discovery through a restricted Redis-compatible protocol. Hub's runtime lock service is separate from Watch and MQ; embedded locking is the default, with external Redis or disabled locking available in network deployments. An application-owned `infra/redis` component manages business Redis access.

## Select a Topology

| Mode | Processes | Evidence it can provide |
| --- | --- | --- |
| Standalone | Hub, Portal, Link, App together | In-process application behavior and any explicitly started real Portal/Admin listeners |
| Linked | External Hub/Portal; Link and App together | Control-plane networking, leases and shared App/Link process failure |
| Separated | Hub, Portal, Link, App separate | App transport, leases, network security, draining and independent failures |

Use standalone for local composition, or separate daemons with `app.New` to validate App-to-Link and inter-App JSON/CBOR transport. Keep an App and its Link API on the same trusted host/Pod.

## CLI and Listener Baseline

```bash
vine version --json
vine hub serve --help
vine link serve --help
vine portal serve --help

# Seed-backed, read-only Hub; embedded MQ and locks are defaults.
vine hub serve --seed-data-file ./src/server/seed/hub.yaml
# Or select writable persistence explicitly on first initialization:
# vine hub serve --db-sqlite-file ./hub.sqlite --seed-data-file ./src/server/seed/hub.yaml

vine portal serve --hub-endpoint http://127.0.0.1:7071
vine link serve --api-listen 127.0.0.1:7079 \
  --ingress-listen 127.0.0.1:7082 --hub-endpoint http://127.0.0.1:7071
VINE_LINK_ENDPOINT=http://127.0.0.1:7079 ./business-app
```

Network Hub defaults: Control `127.0.0.1:7071`, Watch `127.0.0.1:7072`, Admin/Dashboard `127.0.0.1:7099`. Link App API defaults to `127.0.0.1:7079`; pin ingress for production. App HTTP defaults to a dynamic loopback port. Portal starts listeners from enabled entries; Hub owns the Dashboard listener.

Independent daemon environment variables use `VINE_HUB_*`, `VINE_LINK_*`, and `VINE_PORTAL_*`. Hub uses `--watch-listen`, `--mq-mode=embedded|nats`, `--mq-nats-endpoint`, `--lock-mode=embedded|redis|disable`, and `--lock-redis-endpoint`. Standalone flags scope these names with `--hub-`; linked flags use `--link-`. Renamed application flags derive environment names directly from their flag names.

## Standalone, Seed and Dashboard

```go
standalone.NewWithOption[*application.App](standalone.Option{
    HubSeedDataFile: "./seed/hub.yaml",
    HubAdminListen:  "127.0.0.1:7299",
}).StartAndWait()
```

- Set `HubAdminListen` to enable the direct Hub listener for embedded Dashboard and `/api/invoke`; an empty address disables it. Direct vRPC clients send a full trace id and span (vRPC TS `traceMode: "direct"`). Access control relies on loopback isolation or an authenticated management boundary.
- In default no-db mode, Hub requires a seed and loads read-only configuration into memory. Use `HubSeedData: "{}"` for deliberately empty configuration.
- SQLite/PostgreSQL imports seed, variables and source inputs during initial setup. Later starts read stored configuration. Use the supported management APIs for stored edits, or deliberately choose no-db mode for deployment-owned configuration.
- `HubSeedVarsFile` and `HubSeedVars` supply typed variables; explicit null/zero values take precedence over defaults. YAML anchors, aliases, merge keys and ambiguous numeric forms are rejected.
- Keep business RDB and Hub persistence distinct, and preserve both datasets through reviewed migrations and verified backups.

### Current Portal Seed

The server example's [hub.yaml](example-greeting/src/server/seed/hub.yaml) is validated by real Portal tests. In current seed vocabulary:

```yaml
portalEntries:
  - name: local
    protocol: http
    listenIPs: [127.0.0.1]
    http:
      host: ""
      http: ENABLED
      https: DISABLED
      autohttps: DISABLED
      httpPort: 7188
portalSites:
  - name: greeting-rpc
    type: RPCGW
    actorSkelName: greeting.GreetingActor
portalRules:
  - name: greeting-api
    entryName: local
    matchPathPrefix: /api
    routeType: SITE
    routeSiteName: greeting-rpc
```

The resulting invocation prefix is `http://127.0.0.1:7188/api/invoke`. Declare entries explicitly in the same seed and reference them from rules. `enabled` defaults to true. Use entry-owned transport settings and the current rule fields shown above; the upgrade guide maps earlier seed formats.

Current-format entries nest the host under `http` and use enum transport switches `AUTO`, `ENABLED`, `DISABLED`. Empty-host entries serve HTTP IP access only. TLS needs a matching certificate and SNI/Host. One-label wildcard host matching is a WEBGW feature. Portal certificate fields contain raw PEM `certificate` and `privateKey`.

Host/transport selection precedes path matching. Exact hosts precede wildcard and empty-host fallback; path prefixes compete within the selected group. A Web contract's declared mount path controls its effective match and forwarding paths.

## Registration and Graceful Stop

App before-start hooks run before endpoints and registration; requests may arrive before after-start hooks complete. Probe the actual caller path after startup to verify Portal and remote Link convergence.

Link and Portal select endpoints from snapshots. A failure at the selected endpoint returns to the caller, which owns any retry decision. Backend `rpc.WithDestination(appName)` restricts selection to the named App; Portal strips external destination options.

For startup/shutdown hook ordering use [foundations.md](foundations.md). Keep dependencies available until in-flight requests drain. Stop producers and join workers before unregistering; close final resources afterward. Call `StopGracefully` explicitly for graceful shutdown and make running work observe context cancellation.

Validate cross-process heartbeat, leases and App health in a network deployment. Read liveness and drain constants from the target source, then set orchestrator grace accordingly; these constants are version-specific implementation details.

## Dashboard Version Validation

Choose a stable project-specific loopback Admin address. Check port ownership before startup; report a conflict and use an explicitly agreed resolution before continuing. Keep embedded assets and Hub runtime on the same reviewed revision. Clear `VINE_HUB_DASHBOARD_DEV_PROXY` during release validation so Hub serves the bundled Dashboard assets. Set it when developing Hub's frontend proxy.

After a control-plane upgrade, load the real Dashboard and perform an Admin request from its origin. Verify both a fresh session and any reused browser session if caching is relevant. A retired full service name in a response usually means stale generated clients/assets or the wrong running revision. Compare the requested name to the current descriptor and align generated clients, bundled assets and the running revision.

## Security and Durability

Backend mTLS requires all of `--mtls-ca-file`, `--mtls-cert-file`, and `--mtls-key-file`, with the target's exact SPIFFE daemon identities:

```text
spiffe://<trust-domain>/vine/daemon/vine.hub
spiffe://<trust-domain>/vine/daemon/vine.link
spiffe://<trust-domain>/vine/daemon/vine.portal
```

It protects network Control/Watch/MQ and Link ingress paths as implemented by the selected runtime. The browser Admin listener and local App-to-Link API rely on their management and internal access boundaries. Keep those listeners within the intended trust boundaries, and keep public Portal TLS certificates separate from daemon identities.

Embedded NATS provisions memory-backed `VINE_EVENTS` (`event.>`, interest retention) and `VINE_TASKS` (`task.>`, work-queue retention). External NATS deployments must provision both streams before runtime startup and own authentication, TLS, storage and replication. To establish restart durability, choose file-backed streams and verify recovery.

Event/Task delivery is at least once. The application owns deduplication, failure storage and any DLQ. Task failures retry until acknowledgement, using exponential backoff with jitter bounded between roughly one second and ten minutes. Retain business idempotency and durable records/outbox where needed. Deploy the supported single-Hub control plane and plan recovery around it.

## Validate Deployment Changes

Record exact binaries and selected modules. In the target topology validate business API/Web calls, relevant App-to-App traffic, configuration resolution, Actor/trace/deadline propagation, and control-plane access. For changed infrastructure, exercise independent failures, lease expiry, cancellation/draining and restoration of backed-up state. Test external NATS stream configuration and restart recovery when durability is claimed.

Persistent database upgrades, multi-host mTLS, external Redis/NATS, browser-cache behavior and production rollout each require dedicated validation. Record the checks performed and any remaining coverage gaps.
