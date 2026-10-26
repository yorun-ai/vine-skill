# Upgrading from the v0.13.2 Skill Baseline

This is an impact guide for adopting Vine `v0.32.5` with skelc `v0.34.0`. Apply the rows that affect the existing project and preserve compatible components. Source revisions and release notes are indexed in [official-sources.md](official-sources.md).

| Boundary | Required adjustment |
| --- | --- |
| Compiler/module | Migrate the compiler installation from `go.yorun.ai/skelc` to `go.yorun.ai/skel/cmd/skelc@v0.34.0`. Use Go 1.27; this package tests 1.27.3. Regenerate all maintained targets. |
| Generated Go | Backend modules explicitly pin Vine v0.32.5. Portable scalars come from `go.yorun.ai/skel/types`; descriptors from `go.yorun.ai/skel/descriptor`. Migrate UUID constructors from `github.com/google/uuid` to standard-library `uuid`. Current non-nullable nil collections encode as empty arrays/maps; nullable collections retain null. |
| API vs backend | Declare external contracts as `api service ...ApiService`; ordinary `pub service ...Service` generates backend Clients and Servers. Actor audiences and admission belong to API/Web contracts. Backend generation emits API Servers; use `--api` to generate their external clients. Share business implementations through DI. |
| Authentication | Declare `auth required`, `optional` or `anonymous`; Web additionally supports `off`. Methods can inherit the service mode. Retired `auth`/`noauth`/`unset` spellings must be migrated. Invalid credentials are rejected even by `optional`. |
| Actor audiences | Remove `via` clauses and `actorVia` from maintained contracts/seeds. Use `for AActor, BActor`; current audience and descriptor APIs select actors only. |
| Configuration | Structured fields and abstract config adoption are supported. Values preserve string whitespace. Reserve `AbsConfig` for abstract configs and `ApiService` for API services. Register and seed a concrete adopter; resolve abstract config through explicit adoption/binding rules. |
| Runtime helpers | Migrate `core/runtime` uses to `buildinfo` and injected `meta.CurrentApp`. In-process Rpc uses runtime cloning of declared Go values. Regenerate legacy validation hooks and schema/error/codec helpers, then use the generated facades or corresponding public APIs. |
| Standalone | Replace `SQLiteFile`, `SeedYAMLFile` and `DashboardURL` with `HubDBSQLiteFile`, `HubSeedDataFile` and `HubAdminListen`. Set the latter to a listen address such as `127.0.0.1:7299`. Empty Admin listen disables it. Linked options use `LinkHubEndpoint`, `LinkIngressListen`, and `LinkMTLS*`. |
| Hub storage/seed | Default is read-only no-db with a required seed. Persistent databases seed only once. Remove `override`; use explicit entries, current rules and `enabled`. Structured config YAML is accepted in addition to JSON strings. |
| Portal entry | Current entries use `protocol: http` and nested `http.host`, `http`, `https`, `autohttps`, `httpPort`/`httpsPort`; transport switches are `AUTO`, `ENABLED`, `DISABLED`. Rules use `entryName`, `matchPathPrefix`, `routeType`, `routeSiteName`, `routePathPrefix`. |
| Dashboard | Hub's unauthenticated Admin listener serves Dashboard and `/api/invoke` directly. Configure it through `HubAdminListen` and protect access with loopback isolation or authenticated infrastructure. |
| CLI/distribution | Migrate `vine dev` workflows to standalone or separated daemons, and update Redis/MQ/seed flags to the names in the runtime reference. Hub Watch owns the restricted Redis-compatible distribution interface; runtime locks are a separate service. Independent daemon environment variables use `VINE_HUB_*`, `VINE_LINK_*`, `VINE_PORTAL_*`. |
| Web serving | Migrate `web.NewReverseProxy`/`ReverseProxy`/`ProxyOption` uses to `web.DevProxyServer` for a state-file workflow, or `httputil.ReverseProxy` with `web.ServeReverseProxy`. Embed `web.AssetsServer` by value with per-execution context. |
| Timeouts and cancellation | RPCGW still defaults to 30s with a 120s maximum. WebGW supports site/path `webTimeout` policies and most-specific matching; client options can only shorten a configured policy. Portal owns SSE idle enforcement; internal Link proxies preserve cancellation and transport errors. |
| Redis release | `Lock.TryUnlock()` atomically checks local state and token ownership; false means the lease is unavailable/lost/released, while Redis command errors still panic. Replace hand-written `IsBroken`/`Unlock` check-then-act sequences where appropriate. |
| Data/testing | `rdb.Dao.EnsureSchema` is called automatically; implement the hook to perform application-owned schema work. Explicit migrations remain application-owned. `execution.NewClient[T]()`/`NewClientER[T]()` are generic methods for backend services. Test API admission through Portal. |
| External vRPC | Use vRPC Go v0.15.1 and vRPC TS 0.10.0 with generated clients. Shared HTTP codecs own envelopes. RpcGW forwards its fixed request-header contract. Carry application context in declared parameters or supported metadata. |

## Dependency Evidence

Vine v0.32.5 and vRPC Go v0.15.1 declare Skel v0.33.0. The backend module generated by skelc v0.34.0 declares Skel v0.34.0, so the example application selects v0.34.0 through Go minimal version selection. This is intentional: skelc v0.34.0 changes CLI schema resolution/defaults and retains the generated runtime contract from v0.33.0. Verify the selected graph with `go list -m` alongside each dependency's declared requirement. Retain the compatible Vine/vRPC dependency declarations for this Skill update.

`schema get` now resolves imports by default. Pass every required `--skel-import domain=PATH` mapping, or use `--shallow` when opaque imported references are sufficient. `schema list` remains shallow by default.

## Persistent Hub Upgrades

Validate persistent upgrades against a copy of the actual stored database. Vine v0.32.5's automatic migrations support the v0.27/v0.28 storage baseline. A v0.13-era database needs staged migration, backups and restore tests against the intermediate release notes: v0.15.7, then v0.21.5, then v0.27 before the current runtime. Confirm the stored schema, preserve the data, and validate every migration stage on a restored backup before a production upgrade. Rollback requires the matching old binary, seed and database backup.

## Optional Capabilities to Consider

- Use `core/lock.Locker` or `UniversalLocker` when runtime-managed coordination fits the deployment; `redis+memory://` only coordinates within one process.
- `InitHooks` provides injected lifecycle callbacks directly on the App.
- Version-suffixed contracts are independent contracts; client adoption and old-version retirement require explicit migration.
- RPCGW inspection and MCP are opt-in per site. Inspection makes contract metadata publicly readable; enable it for a deliberate inspection use case. MCP tools still use Vine credential admission and execution budgets.

Adopt each capability when its behavior serves a concrete project requirement.
