# Browser Frontends with Generated TypeScript vRPC

Use this reference when a browser frontend calls a Vine business service. Read [foundations.md](foundations.md) for the generation gate and [capabilities.md](capabilities.md) for Rpc/Web selection and vRPC semantics. Example package, domain, and service names below are placeholders; use the target project's actual `package.json`, generated exports, and existing code.

## Required Boundary

Use this path for every structured synchronous browser call:

```text
.skel Rpc → pinned skelc generation → generated Go Server + generated TypeScript service/spec/types
          → project vRPC client → Portal entry/admission → Link → App
```

Deliver the browser application itself through a separate generated Web capability:

```text
.skel Web → pinned skelc check/gen/diff → generated Default...WebServer
          → app.WebberEnabled → WebberInitHandlers → WEBGW site/rule
          → Portal → Link → App Web handler → Vite proxy or production assets
```

- Treat hand-maintained `.skel` as the contract and generated TypeScript as the frontend API boundary.
- Use Rpc/vRPC for CRUD, queries, search, pagination, authentication, and state changes.
- Use Web/HTTP only for multipart, uploads, downloads, Range, `Content-Disposition`, binary streams, or static assets. Use Rpc first for metadata, authorization, and transfer-session creation.
- Never call App or Link internal listeners from the browser. Enter through the project's public Portal origin and admission policy.
- Never forge `vrpc-actor` or `vrpc-initiator`; trusted identity is established by entry policy.

## Discover Before Implementing

1. Confirm the project's pinned Vine and skelc versions and its existing generation command. Do not install or upgrade either tool.
2. Locate the `.skel` source, generated Go output, generated TypeScript package, package exports, generated-file markers, and build scripts.
3. Inspect the generated TypeScript exports for the actual service factory, `ServiceSpec`, method signatures, request/result types, and structured errors. Do not infer names from this document.
4. Locate the project's shared vRPC client, Portal base URL/configuration, authentication integration, and existing React Query or equivalent server-state layer.
5. Locate the server's generated default Server implementation and App capability registration so the frontend method maps to a registered full Skel service/method name.

If the contract or generated boundary changes, modify `.skel`, run pinned `skelc check`, run the repository's generation command for every maintained target, and review both Go and TypeScript diffs before implementing consumers. If the UI only consumes an unchanged method, leave `.skel` and generated files untouched.

## Publish a Local Standalone Browser App on Port 7288

Apply this section only when the task includes a browser frontend and local Standalone delivery. Respect an existing project's explicitly pinned ports and topology. For a new project without such a convention, use this default:

| Address | Responsibility |
| --- | --- |
| `http://127.0.0.1:7288/` | Public browser page through Portal → WEBGW |
| `http://127.0.0.1:7288/api/invoke` | Business vRPC through Portal → RPCGW |
| `http://127.0.0.1:7299/` | Project-specific Standalone Dashboard example; never the business page |
| `http://127.0.0.1:5174/` | Private Vite development upstream |

The page and API deliberately share one public origin. The `/api` and `/` rules share port `7288`; Vine selects the longest matching path prefix, so `/api/invoke` reaches RPCGW before the `/` frontend catch-all. Check both `7288` and the chosen Dashboard port before startup. If either belongs to another process, stop and report it; do not kill the owner or silently choose another port.

### Generate and register the Web capability

Declare the browser asset entry in the hand-maintained Skel contract. The actor and names are placeholders, and authentication policy must come from the target project; do not add `noauth` merely for convenience.

```skel
domain example.console

pub actor BrowserActor {
    via client {}
}

web ConsoleWeb {
    for BrowserActor via client
}
```

Run the repository's pinned `skelc check` and all maintained Go/TypeScript generation targets, then review the diff. Generation must produce the Web schema and `DefaultConsoleWebServer`; do not recreate that boundary by hand.

Embed the Web capability in the App and register the concrete generated handler. Own the development proxy in a Module so App shutdown can close it.

```go
type ConsoleApp struct {
    app.Application
    app.ServicerEnabled
    app.WebberEnabled
}

func (*ConsoleApp) InitModules(add app.TypeAdder) {
    add(app.T[*frontend.Proxy]())
}

func (*ConsoleApp) WebberInitHandlers(add app.TypeAdder) {
    add(app.T[*frontend.ConsoleWeb]())
}
```

The proxy target defaults to Vite but may be overridden by a project-specific environment variable after validating that it is an absolute `http` or `https` origin without credentials, query, fragment, or a non-root path.

```go
type Proxy struct {
    app.BaseModule

    target       *url.URL
    reverseProxy *web.ReverseProxy
}

func (p *Proxy) DIInit() {
    raw := strings.TrimSpace(os.Getenv("APP_FRONTEND_DEV_URL"))
    if raw == "" {
        raw = "http://127.0.0.1:5174"
    }
    target, err := url.Parse(raw)
    if err != nil ||
        (target.Scheme != "http" && target.Scheme != "https") ||
        target.Host == "" || target.User != nil || target.RawQuery != "" ||
        target.Fragment != "" || (target.Path != "" && target.Path != "/") {
        panic("APP_FRONTEND_DEV_URL must be an absolute http or https origin")
    }
    target.Path = ""
    p.target = target
    p.reverseProxy = web.NewReverseProxy(web.ProxyOption{Target: target})
}

func (p *Proxy) AfterAppStop() {
    if p.reverseProxy != nil {
        p.reverseProxy.Close()
    }
}

func (p *Proxy) Serve(ctx *gin.Context) bool {
    return p.reverseProxy != nil && p.reverseProxy.Serve(ctx)
}
```

The concrete Web handler embeds the generated default Server and owns the catch-all route. If the upstream is down, a registered Web endpoint returns an explicit `503`, distinguishing an unavailable frontend process from missing Vine capability registration.

```go
type ConsoleWeb struct {
    skeled.DefaultConsoleWebServer

    Context web.Context `inject:""`
    Proxy   *Proxy      `inject:""`
}

func (h *ConsoleWeb) Routes(router *web.Router) {
    router.ANY("/*path", h.ProxyFrontend)
}

func (h *ConsoleWeb) ProxyFrontend() {
    ctx := h.Context.Gin()
    if h.Proxy.Serve(ctx) {
        return
    }
    ctx.Header("Retry-After", "1")
    ctx.String(http.StatusServiceUnavailable, "frontend development server is unavailable; start it with `cd src/web && pnpm dev`\n")
}
```

### Configure RPCGW and WEBGW on the same Portal listener

Check in the seed and pass it through `standalone.Option.SeedYAMLFile`. `actorSkelName` and especially `webName` must match the generated full Skel names exactly.

```yaml
portalSites:
  - name: console-rpc
    override: true
    type: RPCGW
    actorSkelName: example.console.BrowserActor
    actorVia: client
    cors:
      mode: STRICT
      allowedOrigins:
        - http://127.0.0.1:7288
        - http://localhost:7288
  - name: console-web
    override: true
    type: WEBGW
    actorSkelName: example.console.BrowserActor
    actorVia: client
    webName: example.console.ConsoleWeb
    cors:
      mode: STRICT
      allowedOrigins:
        - http://127.0.0.1:7288
        - http://localhost:7288
portalRules:
  - name: console-api
    override: true
    scheme: http
    host: ""
    port: 7288
    pathPrefix: /api
    targetType: SITE
    siteName: console-rpc
    redirectionPattern: ""
  - name: console-web
    override: true
    scheme: http
    host: ""
    port: 7288
    pathPrefix: /
    targetType: SITE
    siteName: console-web
    redirectionPattern: ""
```

Use a different, project-specific Dashboard origin and load the seed from the Standalone entry point:

```go
standalone.NewWithOption[*application.ConsoleApp](standalone.Option{
    SeedYAMLFile: "./src/server/seed/hub.yaml",
    SQLiteFile:   "./data/hub.sqlite",
    DashboardURL: "http://127.0.0.1:7299/",
}).StartAndWait()
```

Run a non-destructive ownership check immediately before constructing Standalone:

```go
func requireAvailablePort(address string) error {
    listener, err := net.Listen("tcp", address)
    if err != nil {
        return fmt.Errorf("required listener %s is already in use: %w", address, err)
    }
    return listener.Close()
}

func requireAvailablePorts(addresses ...string) {
    for _, address := range addresses {
        if err := requireAvailablePort(address); err != nil {
            log.Fatal(err)
        }
    }
}

// Call immediately before standalone.NewWithOption.
requireAvailablePorts(":7288", ":7299")
```

Adapt fatal-error reporting to the project's `main` shape. This short check has an unavoidable check-to-bind race, so the actual Standalone bind failure must also remain fatal and visible.

### Start and verify the development topology

Every delivered browser-enabled project must include the bundled [Linux Bash launcher](../scripts/start_vine_app.sh) and [Windows Batch launcher](../scripts/start_vine_app.bat) in its own `scripts/` directory. Adapt only project paths and documented environment defaults; keep their preflight, port, readiness, and child-cleanup behavior aligned.

Recommended startup:

```bash
# Linux, first run
./scripts/start_vine_app.sh --install

# Linux, later runs
./scripts/start_vine_app.sh
```

```bat
rem Windows Command Prompt, first run
scripts\start_vine_app.bat --install

rem Windows Command Prompt, later runs
scripts\start_vine_app.bat
```

Enter this section and create `src/web/` only after the user explicitly confirms a frontend. Use `--check` to validate root-level `skel/` and `skeled/`, the standard server directories `src/server/` and `src/server/seed/hub.yaml`, the frontend package `src/web/`, `go`/`pnpm`, frontend dependencies, and port availability without starting any process. The scripts require only the platform shell plus the project's existing Go and pnpm prerequisites; they must not require Python. If a port is occupied, stop and report it; do not terminate the existing process. Set `VINE_PREPARE_PACKAGE=./src/server/cmd/migrate` if and only if the project explicitly contains a migration entry and requires a preparation command; never infer a migration from directory names.

A delivery reply cannot just say "the frontend is done" or point the user at the README. The final reply must directly provide:

- The Go, Node.js, pnpm, and other prerequisite versions required by the project.
- The Linux/macOS first-run and later-run startup commands, and the corresponding Windows Command Prompt commands.
- The combined launcher's `--check` preflight-only command.
- The complete access addresses for the business page, public vRPC, and Dashboard, and a clear statement that the private Vite upstream is not a user entry.
- The install entry to run when dependencies are missing, and, when a port is occupied, the checked port and its owning process. Never terminate an existing process or silently switch ports.
- When the project has migrations, whether the combined launcher runs them automatically or they must be run manually first; do not invent migration steps when no migration entry exists.

The manual fallback is two terminals:

```bash
# Terminal 1
cd src/web
pnpm install
pnpm dev

# Terminal 2, from the project root
go run ./src/server/cmd/demo
```

Replace the example entry `demo` with the real project name. Do not move existing server directories when adding the frontend.

Visit `http://127.0.0.1:7288/`, not the Vite upstream or Dashboard. Configure the shared vRPC transport once with `http://127.0.0.1:7288/api/invoke` (or the equivalent same-origin `/api/invoke` value); feature code still calls only generated services. If direct Vite-origin browsing is also supported, add a Vite development proxy for `/api/invoke` and the corresponding strict CORS origin rather than changing business calls.

Pin the Vite development listener instead of allowing an automatic fallback to another port that the Go proxy does not know about:

```ts
export default defineConfig({
  server: {
    port: 5174,
    strictPort: true,
  },
})
```

For production, build the frontend and serve the output from the generated Web handler with public Vine APIs such as `web.NewEmbedAssetsAccessor` and `web.NewAssetsServer`. Do not run or reverse-proxy a Vite development server in production. A minimal per-request serving shape is:

```go
//go:embed dist
var frontendFS embed.FS

var frontendAssets = web.NewEmbedAssetsAccessor(frontendFS, "dist")

func (h *ConsoleWeb) ServeFrontend() {
    assets := web.NewAssetsServer(frontendAssets)
    assets.ServeAsset(h.Context.Gin(), frontendAssets)
}
```

Keep `ANY("/*path")` or the generated route appropriate to the target revision so client-side routes fall back to the built `index.html` through the assets server.

### Diagnose WEBGW failures

`webgw endpoint is unavailable: <site>` has a specific meaning: the seed successfully selected a WEBGW site, but the configured full `webName` is not registered by an App on the selected Link. Check, in order:

1. the `.skel` Web declaration exists and pinned generation produced its Go Web boundary;
2. the seed's `webName` exactly equals the generated full Skel Web name;
3. the App embeds `app.WebberEnabled`;
4. `WebberInitHandlers` registers the concrete handler embedding `Default...WebServer`;
5. Portal, Link, and App registration have converged; and
6. the running process is the newly built process rather than an old instance.

Do not attribute this error to Vite. Once the Vine Web endpoint is registered, an unavailable Vite upstream must reach the handler and return the deliberate `503 Service Unavailable` plus `Retry-After: 1`. A bare `404` at `:7288/` usually means the `/` WEBGW rule/site is absent or does not match the request.

## Use the Generated Service

Create the generated service from the project's shared vRPC client outside React render code. Keep business calls in feature-local query hooks; pages consume hook results and mutations.

```ts
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { vrpcClient } from '@example-app/web-base'
import {
  CatalogReportServiceSpec,
  createCatalogReportService,
} from '@example-app/skeled-catalog'

const catalogReportService = createCatalogReportService(vrpcClient)

export const catalogReportKeys = {
  listRoot: () => [
    CatalogReportServiceSpec.serviceName,
    CatalogReportServiceSpec.methods.list,
  ] as const,
  list: (status: string) => [...catalogReportKeys.listRoot(), { status }] as const,
}

export function useCatalogReports(status: string) {
  return useQuery({
    queryKey: catalogReportKeys.list(status),
    queryFn: () => catalogReportService.list({ status }),
  })
}

export function useArchiveCatalogReport() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationKey: [
      CatalogReportServiceSpec.serviceName,
      CatalogReportServiceSpec.methods.archive,
    ],
    mutationFn: (id: string) => catalogReportService.archive({ id }),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: catalogReportKeys.listRoot(),
    }),
  })
}
```

Adapt method arguments only after reading the generated signatures. Import generated types when an explicit annotation is needed; do not copy their fields into parallel frontend DTOs. Derive query/mutation identity from the generated `ServiceSpec`, then append stable, serializable resource IDs or normalized filters. Preserve generated business/SystemError distinctions instead of flattening every failure into an empty result.

Do not:

- hand-write vRPC URL paths, media types, headers, envelopes, service names, or method names for ordinary product code;
- use `fetch`, Axios, or a JSON REST route for a structured business Rpc;
- edit generated TypeScript or add a hand-maintained client beside it;
- instantiate a different raw vRPC transport in each feature;
- put functions, component instances, unnormalized `Date` values, or transient objects in query keys.

## Validate the Browser Path

1. For a contract change, retain successful pinned `skelc check`, every maintained Go/TypeScript generation target, and generated diff evidence. Confirm generated files were not hand-edited and no `go.yorun.ai/vine/internal/*` import was added.
2. Unit-test the Web handler with an HTTP test upstream: verify proxy status/body/path, upstream-unavailable `503`, `Retry-After: 1`, and proxy closure during App shutdown. For production assets, test `/`, a static asset, a client-side route fallback, missing assets, and unsupported methods.
3. Start Standalone on dynamically allocated Portal and Dashboard ports using a temporary copy of the real checked-in seed. Verify `Portal → WEBGW → Link → App Web → temporary frontend` returns `200`, and verify the `/api` rule wins over the `/` rule. Do not hard-code shared test ports or replace the real seed with a weaker test-only topology.
4. Run the affected frontend package's typecheck, tests, lint, and production build according to repository scripts.
5. Call at least one affected method through the generated TypeScript service against the real Portal/vRPC origin. Verify success, one relevant business/error path, authentication/Actor policy, and status mapping.
6. Inspect browser requests to confirm the page and business calls use the public Portal origin rather than App/Link listeners or the private Vite upstream, and that the generated full Web/service/method identities are registered.
7. Exercise mutation invalidation and loading/error states. Do not treat a rendered static page, a mocked transport, or a direct Vite response as end-to-end evidence.
8. If Portal reports `rpc service schema is not found`, compare the generated full Skel service name with the schema registered by the target Vine revision. If it reports `webgw endpoint is unavailable`, compare the generated full Web name with the WEBGW site and App registration. Fix generation, registration, origin, assets, or cache rather than adding aliases.
