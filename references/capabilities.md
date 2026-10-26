# Vine Rpc, Web, Event, and Task

This reference is based on Vine `v0.32.5`. Use it when designing or implementing typed capabilities and vRPC HTTP boundaries. Read the version and generated-code rules in [foundations.md](foundations.md) first. The timeout, concurrency, and other values here are `v0.32.5` version facts; the authoritative list is in [Version baseline quick reference](foundations.md#version-baseline-quick-reference).

## Contents

- [Select a capability](#select-a-capability)
- [Rpc](#rpc)
- [Web](#web)
- [Event and Task](#event-and-task)
- [vRPC over HTTP](#vrpc-over-http)
- [Validate capabilities](#validate-capabilities)

## Select a Capability

| Semantics | Choice |
| --- | --- |
| Collaboration inside one App | Plain Go interface + DI |
| Structured synchronous business request/response across Apps, browsers, or external clients | Rpc/vRPC |
| File upload/download, other binary streams, or static asset delivery | Web/HTTP |
| Publish a fact to every logical consuming App | Event |
| Ask one globally available Runner to perform work | Task |

Rpc/vRPC is the default for structured business APIs; Web handles files, assets and required HTTP or streaming protocols. Keep application-owned CRUD, queries, search, pagination, authentication, and state changes on Skel Rpc through Portal's vRPC entry. Use Web for multipart, uploads, downloads, Range, `Content-Disposition`, byte streams, SSE/WebSocket, static assets, and integrations that require an HTTP protocol.

Ask first: Does this boundary need independent deployment? Does the caller wait for a result? Is one logical delivery needed or several? Can failures retry? Does the business need a durable record? Use the generated Rpc/vRPC boundary for structured synchronous calls.

Shared workflow:

1. Split Services by business entity, aggregate, or named use case first, then declare types, identity, and capabilities completely in an independent `.skel` file. Name each boundary after its entity, aggregate or use case.
2. Run `check` and every maintained Go/TypeScript `gen` target with the pinned skelc, then review the generated boundary.
3. Only after generation succeeds, embed `Default...` and write the concrete implementation outside the generated package.
4. Embed the corresponding `...Enabled` in the App and register the handler, listener, or runner.
5. Inject generated Go clients, emitters, or launchers into Go callers; browser callers import the generated TypeScript service/spec/types and construct the service with the project vRPC client.
6. Validate behavior with testkit and validate network/runtime semantics in the target topology.

## Rpc

For two domains consuming each other's generated Clients and Event Listeners, follow
[Orders and Catalog](example-domains/README.md). It includes public/regular package generation,
an imported DTO, separate App registrations and a tested request/Event round trip.

Separate `pub service ...Service` backend contracts from `api service ...ApiService` external contracts. API services are reached through Portal. Backend generation emits their Servers; use `--api` for external Go/TypeScript clients. Actor audiences and permission admission belong to API/Web contracts. Share business logic through DI when both entry points need it.

API services and Web contracts declare explicit `auth required`, `optional` or `anonymous`; Web alone also supports `off`. Methods inherit the service policy unless explicitly overridden. Use `for ClientActor` or a comma-separated actor list to select the audience.

| Mode | Missing credential | Supplied credential |
| --- | --- | --- |
| `required` | Reject | Authenticate; reject invalid credentials |
| `optional` | Anonymous | Authenticate; reject invalid credentials |
| `anonymous` | Anonymous | Requires an anonymous caller; rejects supplied invalid or authenticated credentials |
| Web `off` | Handler owns auth | Preserve Authorization for native handler authentication |

Portal strips Authorization after Vine admission for the other modes. Call an optional login endpoint with credentials omitted for anonymous admission, or with a valid credential for authenticated admission. Migrate earlier `noauth`, `auth` and `unset` spellings to one of the explicit modes above.

An App can register multiple Rpc Services. A multi-entity App must use concrete boundary names, such as `CustomerService`, `DealService`, `ActivityService`, and `DashboardService` in a CRM App. The corresponding `src/server/impl` Handlers, `src/server/core` Services, and `src/server/repo` Repositories are also split by the same entity or use case.

For a backend `pub service GreetingService`, the implementation is:

```go
type GreetingService struct {
    skeled.DefaultGreetingServiceServer
}

func (*GreetingService) Hello(name string) skeled.Greeting {
    return skeled.Greeting{Message: "Hello, " + name}
}

type GreetingApp struct {
    app.Application
    app.ServicerEnabled
}

func (*GreetingApp) ServicerInitHandlers(add app.TypeAdder) {
    add(app.T[*GreetingService]())
}
```

Inject the generated client directly into the caller:

```go
type Probe struct {
    Client skeled.GreetingServiceClient `inject:""`
}
```

Rules:

- An App owns assembly and runtime lifecycle. `ServicerInitHandlers` should register every concrete entity/use-case Handler.
- Each generated Service, default Server, `src/server/impl` Handler, `src/server/core` Service, and `src/server/repo` Repository must keep traceable names and responsibilities. Services and Repositories may share infrastructure while retaining their entity/use-case responsibilities.
- Use generated `ServiceSpec`, `MethodSpec`, schemas and registration in business code.
- Server implementations outside the generated package must embed the generated default Server to satisfy the sealed interface.
- A client resolved inside an execution inherits the current trace, Actor, Initiator, and deadline. A client injected into a Module uses the App root context.
- The default call timeout in `v0.32.5` is 30 seconds. Use generated capabilities or public invocation options when another value is needed, and verify the target version.
- Low-level `WithContext` changes the Go parent context and preserves the existing trace, Actor and Initiator. Choose either `WithContext` or `WithTimeout` for an invocation.
- SystemError panics by default. When explicit error returns are required, use the generated ER client or a public option. Preserve the structured failure classification across the caller boundary.
- In-process Rpc clones declared Go argument/result types directly in the runtime. `MethodSpec.CloneArguments` and `CloneResult` remain deprecated declarations; runtime reflection owns the copying operation. Use generated scalars, collections and beans to preserve value isolation.
- Use in-process Rpc tests to verify value isolation. Use a separated topology to test JSON/CBOR encoding, transport normalization, custom marshal behavior and codec failures.
- Link chooses one available endpoint per call using round-robin selection and returns that endpoint's result or failure. Implement retries, affinity, weighted routing, circuit breaking or pre-call health probes in an explicit higher-level design when required.
- Design explicit retries only when the call is idempotent and enough deadline budget remains.

## Web

Use Web for binary/asset transfer and required HTTP/streaming protocols, such as:

- Multipart file uploads.
- Large file downloads, Range, and `Content-Disposition`.
- Streaming binary responses, SSE and WebSocket.
- Static frontend assets.

Declare CRUD, search, pagination, and action commands for structured product, order, user, or permission data as Rpc, and have frontends or external clients call them through vRPC over HTTP. If one use case contains both structured metadata and file bytes, use Rpc/vRPC first for metadata, authorization, or an upload session, then use Web for the bytes, bound to the same Actor and business ID.

Rules:

- An App has only one Webber; register multiple generated Web Handlers within it.
- A Handler method uses an empty parameter list and obtains its request context and dependencies from execution DI.
- `web.Router` supports common HTTP methods and subrouters. Use `web.NewAssetsServer` for static resources.
- A local browser frontend must still declare and generate its Skel Web boundary, embed `app.WebberEnabled`, register the handler through `WebberInitHandlers`, and expose the exact full Web name through a Portal WEBGW site. During development, the registered handler may serve Vite through `web.DevProxyServer` or a lifecycle-owned `httputil.ReverseProxy` with `web.ServeReverseProxy`.
- Keep application-owned business DTOs on Skel Rpc/vRPC. A required HTTP integration, such as a webhook, uses its mandated JSON payload at that integration boundary.
- Public business traffic must pass through Portal entry/site/admission and then be delivered to the local App through Link webproxy.
- Restrict direct App Web endpoints to trusted internal callers; they consume headers written by Vine.
- Use Rpc/vRPC for structured health, readiness, and business probe results as well.

For the same-origin `:7288` RPCGW + WEBGW Standalone pattern, startup, production assets, and `webgw endpoint is unavailable` diagnosis, use [frontend-vrpc.md](frontend-vrpc.md#publish-a-local-standalone-browser-app-on-port-7288).

## Event and Task

For a runnable contract → generated emitter → Listener → generated launcher → Runner flow,
see [Greeting's asynchronous receipts](example-greeting/README.md#event-to-task-receipts).
The example includes App registration, SQLite deduplication, a result-query API and real delivery/retry tests.

### Consumer Semantics

| Question | Event | Task |
| --- | --- | --- |
| Meaning | A fact that already occurred | A request to perform work |
| Logical delivery | One copy for each distinct listening App name | One copy globally |
| Replica behavior | Instances with the same App name compete for that App's copy | All Runners compete for the same work |
| Sender waits for | Link to accept the publish | Link to accept the publish |
| Result delivery | Receiver handles asynchronously | Receiver handles asynchronously |

A Task is claimed by one eligible Runner across the deployment. When two logical consumers must both observe an Event, use Listeners with different App names. When horizontally scaling one logical consumer, share the same App name.

### Registration Options

```go
app.WithListenerTimeout(20*time.Second)
app.WithListenerConcurrency(4)
app.WithListenerNoRetry()

app.WithRunnerTimeout(10*time.Minute)
app.WithRunnerConcurrency(2)
app.WithRunnerNoRetry()
app.WithRunnerCronScheduler("nightly", "0 2 * * *")
```

Defaults and limits in `v0.32.5`:

| Behavior | Event Listener | Task Runner |
| --- | --- | --- |
| Per-attempt timeout | 30 seconds | 30 seconds |
| Concurrency per registration instance | 10 | 10 |
| Failure | Retry | Retry |
| Retry policy | Continue until acknowledgement | Continue until acknowledgement |
| `NoRetry` | Terminal acknowledgement; application owns failure records | Terminal acknowledgement; application owns failure records |
| DLQ ownership | Application | Application |
| Stream storage | Embedded NATS: memory; external NATS: deployment-owned | Embedded NATS: memory; external NATS: deployment-owned |

These are `v0.32.5` version facts (see [Version baseline quick reference](foundations.md#version-baseline-quick-reference) for the full list). Check the corresponding documentation and source when the target version differs.

### At-Least-Once Delivery and Cancellation

- Put a stable `eventId` or `jobId` in the contract. Protect side effects with unique constraints, transactional deduplication, idempotent upserts, or external idempotency keys.
- Lost acknowledgements, errors, timeouts, reconnects, or consumer reassignments can all cause duplicate execution.
- `NoRetry` acknowledges a failed attempt as terminal. Persist failure records or DLQ entries explicitly when the business needs recovery or inspection.
- A timeout cancels one attempt's context. Handlers must observe cancellation to stop; a still-running old attempt may overlap a retry.
- Implement global ordering, replay for late Listeners, audit history and exactly-once business effects through an explicit application protocol when required.
- Embedded NATS uses memory streams. For restart durability with external NATS, provision file-backed streams and verify recovery. For business data requiring durable recovery, keep a database outbox, durable workflow, or another durable business record even when NATS is file-backed.

Event/Task propagates trace, sender App and publication time. Each receiver starts with absent Actor/Initiator metadata and its own deadline and cancellation. For asynchronous authorization, include the minimum immutable identity in the contract and have the receiver revalidate it. Keep credentials and secrets in the owning authentication system.

### Cron

`WithRunnerCronScheduler`:

- Accepts a trigger with an empty input contract.
- Uses the trigger's Skel name.
- Uses standard five-field cron syntax.
- Uses the Hub clock and default local timezone; a `CRON_TZ=Asia/Shanghai` prefix can select a named timezone.
- Schedules work while the runtime is active; missed-run recovery requires an application-owned scheduler or workflow policy.
- Enqueues scheduled messages in the same global Task queue with the same retry and idempotency semantics.

When calendars, misfire policies, execution history, or manual replay are required, use a dedicated scheduler or workflow system and have it launch Vine Tasks.

## vRPC over HTTP

Use Portal's vRPC over HTTP when browsers, mobile applications, external clients, or gateways access structured synchronous business APIs. Ordinary Go applications should prefer generated clients. Handle this wire protocol directly only when implementing a non-Go client, integrating a gateway, or troubleshooting at the wire level.

Request:

```text
POST <prefix>/<service-skel-name>/<method-skel-name>
```

Wire details in `v0.32.5`:

- Set the media type to `application/vrpc+json` or `application/vrpc+cbor`.
- Required request headers: `accept`, `content-type`, `vrpc-trace`, and `vrpc-client`.
- Optional headers: `vrpc-options`, `vrpc-actor`, and `vrpc-initiator`. External entries derive trusted Actor and Initiator values from admission policy.
- Current non-nullable nil lists/maps encode as empty arrays/maps; nullable values retain null. JSON uses the shared codec based on `encoding/json/v2`, and duplicate CBOR map keys are rejected.
- JSON body shape: `{"params": {...}}`.
- Response body: `{"result": ..., "error": ...}`. Interpret the logical outcome using `vrpc-status` and `error.code`.
- The internal App handler's HTTP transport may always return 200 while storing logical status in `vrpc-status`. Portal maps this to conventional external 4xx/5xx statuses.
- Direct transport requires a trace ID and span. Portal may accept an ID-only trace and create the entry span.
- RPCGW defaults to 30 seconds, rejects timeouts above 120 seconds, clears external error details, and applies admission and CORS. It forwards the fixed vRPC header contract and strips external destination options. Carry application context in declared parameters or supported metadata.

Read `vrpc-status` and `error.code` for the logical outcome, and emit each protocol header as a single line. Wire headers, media types, and JSON/CBOR fields are cross-component protocol. When changing them, update every producer, consumer, test, and compatibility note together.

## Optional RPCGW Inspection and MCP

Site `inspectEnabled` makes API contract metadata publicly readable; enable it for a deliberate inspection use case. `mcpEnabled` exposes the site's API tools through `/mcp`; clients provide Vine Authorization credentials directly. The baseline supports MCP protocol `2026-07-28`; tool execution shares Rpc admission, shutdown and timeout policy, with timeout options in `params._meta["vmcp-options"]`. Enable these features when the project requires contract inspection or MCP tool access.

## Validate Capabilities

- Use generated clients to validate Rpc success, business errors, SystemError, timeouts, and context propagation.
- Through Portal, validate admission, Actor, trace, and status mapping for every public Web/Rpc entry.
- Force Event/Task errors, timeouts, duplicate delivery, and cancellation, then verify idempotent state.
- Validate real consumer-group semantics across multiple App names and multiple replicas.
- Use standalone/testkit for business behavior and in-process value isolation. Use separated processes for JSON/CBOR wire behavior, and separated processes for NATS, Link, leases, and networking.
- Pin request/response golden tests for wire integrations, including JSON and CBOR, header validation, and error mapping.

See [execution-and-boundaries.md](execution-and-boundaries.md) for context and error rules, and [official-sources.md](official-sources.md) for official sources.
