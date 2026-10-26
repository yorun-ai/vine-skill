# Browser Frontends with Generated TypeScript vRPC

Use this reference only when a browser frontend is requested or confirmed. It targets Vine `v0.32.5`, skelc `v0.34.0` and `@yorun-ai/vrpc` `0.10.0`. Preserve an existing project's layout, ports and build workflow.

## Required Boundary

```text
api service → generated TypeScript client → Portal RPCGW → Link → App API Server
web contract → generated Web handler → Portal WEBGW → Link → proxy/assets
```

Generate external clients with `skelc gen ts --api --actor <full-actor-name>`. Read generated exports and method arguments before wiring the UI; methods with empty input may take an explicit `null` parameter. Configure one shared transport and use generated services in feature query/mutation hooks. Use the generated DTOs and let the vRPC runtime encode request envelopes.

Public API and Web auth modes must be deliberate. The following asset Web is `optional` so an unauthenticated browser can load the login page. Its API service can independently require credentials. `optional` still rejects invalid supplied credentials; use Web `off` only when the handler owns native authentication.

## Generate a Web Boundary

This complete contract is also compiled by the package's reference validation:

```skel title="browser.skel"
domain example.console

pub actor BrowserActor {}

web ConsoleWeb {
    for BrowserActor
    auth optional
    mount /
}
```

Put the canonical domain declaration in `domain.skel`, and keep the bare domain header in `browser.skel`. Run pinned check/generation for backend and maintained clients. Register the handler on an App embedding `app.WebberEnabled`:

```go
func (*ConsoleApp) InitModules(add app.TypeAdder) {
    add(app.T[*frontend.Proxy]())
}

func (*ConsoleApp) WebberInitHandlers(add app.TypeAdder) {
    add(app.T[*frontend.ConsoleWeb]())
}
```

## Development Proxy

For a fixed Vite upstream, use `httputil.ReverseProxy` with `web.ServeReverseProxy` for Vine completion/error handling. This example owns the transport in a Module and preserves cancellation and proxy response/error behavior. Replace its module import and configured upstream for the target project.

```go title="dev.go"
package frontend

import (
    "net/http"
    "net/http/httputil"
    "net/url"

    "github.com/gin-gonic/gin"
    "go.yorun.ai/vine/app"
    "go.yorun.ai/vine/core/logger"
    "go.yorun.ai/vine/core/web"
    skeled "example.com/console/skeled/golang"
)

type Proxy struct {
    app.BaseModule
    proxy *httputil.ReverseProxy
    transport *http.Transport
    log *logger.Logger
}

func (p *Proxy) DIInit() {
    target, err := url.Parse("http://127.0.0.1:5174")
    if err != nil { panic(err) }
    p.transport = http.DefaultTransport.(*http.Transport).Clone()
    p.proxy = httputil.NewSingleHostReverseProxy(target)
    p.proxy.Transport = p.transport
    p.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        w.Header().Set("Retry-After", "1")
        http.Error(w, "frontend development server is unavailable", http.StatusServiceUnavailable)
    }
    p.log = logger.New("frontend", "proxy")
}

func (p *Proxy) AfterAppStop() { p.transport.CloseIdleConnections() }

func (p *Proxy) Serve(ctx *gin.Context) {
    web.ServeReverseProxy(p.proxy, ctx, p.log)
}

type ConsoleWeb struct {
    skeled.DefaultConsoleWebServer
    Context web.Context `inject:""`
    Proxy *Proxy `inject:""`
}

func (h *ConsoleWeb) Routes(router *web.Router) {
    router.ANY("/*path", h.ProxyFrontend)
}

func (h *ConsoleWeb) ProxyFrontend() { h.Proxy.Serve(h.Context.Gin()) }
```

For an existing state-file development workflow, `web.DevProxyServer` is another supported choice: embed it by value, configure its state file in `DIInit`, and delegate `Routes`. Its followed server outlives one request; an unanswered forwarded request returns 502.

## Production Assets

Use a production-only handler in place of the development handler (for example, mutually exclusive build-tagged files). Embed `web.AssetsServer` by value so its injected context belongs to each execution, and delegate its routes. Place the actual frontend build output in `dist/` beside this file.

```go title="assets.go"
package frontend

import (
    "embed"
    "go.yorun.ai/vine/core/web"
    skeled "example.com/console/skeled/golang"
)

//go:embed dist
var frontendFS embed.FS

var frontendAssets = web.NewEmbedAssetsAccessor(frontendFS, "dist")

type ConsoleWeb struct {
    skeled.DefaultConsoleWebServer
    web.AssetsServer
}

func (h *ConsoleWeb) DIInit() { h.SetAccessor(frontendAssets) }
func (h *ConsoleWeb) Routes(router *web.Router) { h.AssetsServer.Routes(router) }
```

Client-side route fallback depends on an HTML-accepting request; missing non-HTML assets remain 404. Production serves the built assets through the generated Web handler.

## Publish a Local Standalone Browser App on Port 7288

Default addresses for a new project:

| Address | Purpose |
| --- | --- |
| `http://127.0.0.1:7288/` | Business page through Portal WEBGW |
| `http://127.0.0.1:7288/api/invoke` | Business API through RPCGW |
| `http://127.0.0.1:7299/` | Hub Dashboard/Admin, independent of Portal |
| `http://127.0.0.1:5174/` | Private Vite upstream |

The API and page share a public origin. This seed uses the generated Web's `mount /`; its API site assumes the project's API services declare `for BrowserActor`.

```yaml title="browser-hub.yaml"
portalEntries:
  - name: console-local
    protocol: http
    listenIPs: [127.0.0.1]
    http:
      host: ""
      http: ENABLED
      https: DISABLED
      autohttps: DISABLED
      httpPort: 7288
portalSites:
  - name: console-rpc
    type: RPCGW
    actorSkelName: example.console.BrowserActor
  - name: console-web
    type: WEBGW
    actorSkelName: example.console.BrowserActor
    webName: example.console.ConsoleWeb
portalRules:
  - name: console-api
    entryName: console-local
    matchPathPrefix: /api
    routeType: SITE
    routeSiteName: console-rpc
  - name: console-web
    entryName: console-local
    routeType: SITE
    routeSiteName: console-web
```

Use `HubSeedDataFile: "./seed/hub.yaml"` and `HubAdminListen: "127.0.0.1:7299"`. No-db mode reads the seed each start. Persistent Hub storage imports seed during initial setup; apply later changes through the management API or a reviewed migration. The Web's mount sets effective rule prefixes; within this entry `/api` wins over `/`. Use `matchPathPrefix` for rule matching and `HubAdminListen` for Dashboard serving.

Inspect port ownership before startup. Set Vite `server.host: '127.0.0.1'`, `port: 5174` and `strictPort: true`. If an address is occupied, report the conflict and resolve it with the user before startup. Optional launcher templates are [start_vine_app.sh](../scripts/start_vine_app.sh) and [start_vine_app.bat](../scripts/start_vine_app.bat). Their environment port values control launcher checks/URLs; align the Go proxy, seed and Admin address separately.

```bash
# After copying/adapting launchers into the target browser project:
bash scripts/start_vine_app.sh --install
bash scripts/start_vine_app.sh --check
bash scripts/start_vine_app.sh
```

```bat
scripts\start_vine_app.bat --install
scripts\start_vine_app.bat --check
scripts\start_vine_app.bat
```

They expect the documented `src/server/`, `src/web/`, seed and generated-module layout. Install the pinned Go/skelc toolchain and generate contracts before using the launchers. Set `VINE_PREPARE_PACKAGE` only for a real, requested preparation command. Manual startup uses `pnpm dev` in `src/web` and `go run ./cmd/<name>` in `src/server` in separate terminals.

## Use the Generated Service

Create one `createVrpcClient` with the public `/api/invoke` prefix and valid `clientInfo`, then instantiate generated factories such as `createGreetingApiService(client)`. Read exact exports from generated files. With vRPC TS 0.10.0, UUIDs remain strings, nullable fields use `null`, and binary contracts require the appropriate wire schema/CBOR integration from the generated client and runtime.

Keep calls in feature-local query hooks. Derive query keys from the generated `ServiceSpec` plus normalized parameters, and invalidate the matching keys after a mutation. Preserve structured business/SystemError distinctions. Send user credentials through the public authentication contract; Portal establishes internal Actor/Initiator headers.

## Validate the Browser Path

- Compile generated handlers and clients. Test proxy path/status/body forwarding, upstream failure, cancellation and cleanup. For assets test `/`, an HTML route fallback, missing assets and unsupported methods.
- On isolated loopback ports, verify the real seed routes a WEBGW page, that `/api` wins, and that a generated API client succeeds through Portal. Test a relevant admission/error case.
- Run the frontend package's typecheck, tests, lint and build. Inspect the browser page through the public Portal origin, including its generated API calls and credential/CORS behavior.
- `webgw endpoint is unavailable` reports a missing Web endpoint registration for the selected site. Check the full `webName`, generation, `WebberEnabled`, handler registration and convergence before debugging Vite.
- When delivering the app, give prerequisite versions, supported-platform startup/preflight commands, public page/API and Admin addresses, and any remaining startup issue.
