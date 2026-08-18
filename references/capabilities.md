# Vine Rpc, Web, Event, and Task

This reference is based on Vine `v0.13.1`. Use it when designing or implementing typed capabilities and vRPC HTTP boundaries. Read the version and generated-code rules in [foundations.md](foundations.md) first. The timeout, concurrency, and other values here are `v0.13.1` version facts; the authoritative list is in [Version baseline quick reference](foundations.md#version-baseline-quick-reference).

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

Rpc/vRPC is the default for synchronous business APIs; Web/HTTP is the binary-transfer exception. Do not use Web for JSON CRUD, queries, search, pagination, authentication, or state changes. Browsers and external clients should invoke Skel Rpc through Portal's vRPC entry. Use Web only for payloads that do not fit typed messages, such as multipart, uploads, downloads, Range, `Content-Disposition`, large byte streams, or static assets.

Ask first: Does this boundary need independent deployment? Does the caller wait for a result? Is one logical delivery needed or several? Can failures retry? Does the business need a durable record? For structured synchronous calls, do not deviate from Rpc/vRPC merely because REST is familiar or convenient for a frontend.

Shared workflow:

1. Split Services by business entity, aggregate, or named use case first, then declare types, identity, and capabilities completely in an independent `.skel` file. The App name cannot replace a business boundary name.
2. Run `check` and every maintained Go/TypeScript `gen` target with the pinned skelc, then review the generated boundary.
3. Only after generation succeeds, embed `Default...` and write the concrete implementation outside the generated package.
4. Embed the corresponding `...Enabled` in the App and register the handler, listener, or runner.
5. Inject generated Go clients, emitters, or launchers into Go callers; browser callers import the generated TypeScript service/spec/types and construct the service with the project vRPC client.
6. Validate behavior with testkit and validate network/runtime semantics in the target topology.

## Rpc

An App can register multiple Rpc Services. A multi-entity App must use concrete boundary names, such as `CustomerService`, `DealService`, `ActivityService`, and `DashboardService` in a CRM App. Do not create one `CRMService` that merges every method. The corresponding `src/server/impl` Handlers, `src/server/core` Services, and `src/server/repo` Repositories are also split by the same entity or use case.

Standard implementation:

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

- An App is an assembly and runtime boundary, not a default Rpc Service. `ServicerInitHandlers` should register every concrete entity/use-case Handler.
- Each generated Service, default Server, `src/server/impl` Handler, `src/server/core` Service, and `src/server/repo` Repository must keep traceable names and responsibilities. Sharing a database connection is not a reason to merge Services or Repositories.
- Generated code owns `ServiceSpec`, `MethodSpec`, schemas, and registration. Ordinary business code should not recreate low-level specs by hand.
- Server implementations outside the generated package must embed the generated default Server to satisfy the sealed interface.
- A client resolved inside an execution inherits the current trace, Actor, Initiator, and deadline. A client injected into a Module uses the App root context.
- The default call timeout in `v0.13.1` is 30 seconds. Use generated capabilities or public invocation options when another value is needed, and verify the target version.
- Low-level `WithContext` changes the Go parent context but not trace, Actor, or Initiator. It is mutually exclusive with `WithTimeout`.
- SystemError panics by default. When explicit error returns are required, use the generated ER client or a public option. Do not flatten every failure into an empty business value.
- In Vine `v0.13.1`, MethodSpecs generated by skelc `v0.12.0` or later structurally clone in-process Rpc arguments and results. Older generated specs remain compatible through serialization-based cloning. Both paths preserve caller/handler value isolation.
- In-process Rpc guarantees value isolation, not JSON/CBOR encoding, transport normalization, custom marshal behavior, or codec failure handling. Use `vine dev` or a separated topology when those wire behaviors are under test.
- Link round-robins across available endpoints, but it does not provide same-request automatic retry/failover, affinity, weight, circuit breaking, or pre-call health probes. Failure of the selected endpoint returns from the current call.
- Design explicit retries only when the call is idempotent and enough deadline budget remains.

## Web

Web is not a replacement for structured business APIs. Use it only for binary or static asset transfer, such as:

- Multipart file uploads.
- Large file downloads, Range, and `Content-Disposition`.
- Streaming binary responses.
- Static frontend assets.

Declare CRUD, search, pagination, and action commands for structured product, order, user, or permission data as Rpc, and have frontends or external clients call them through vRPC over HTTP. If one use case contains both structured metadata and file bytes, use Rpc/vRPC first for metadata, authorization, or an upload session, then use Web for the bytes, bound to the same Actor and business ID.

Rules:

- An App has only one Webber; register multiple generated Web Handlers within it.
- A Handler method has no parameters and obtains the request context and dependencies from execution DI.
- `web.Router` supports common HTTP methods and subrouters. Use `web.NewAssetsServer` for static resources.
- A local browser frontend must still declare and generate its Skel Web boundary, embed `app.WebberEnabled`, register the handler through `WebberInitHandlers`, and expose the exact full Web name through a Portal WEBGW site. During development, a lifecycle-owned `web.NewReverseProxy` may serve Vite; this does not replace capability registration.
- Web routes must not accept or return structured JSON business DTOs. Replace such routes with Skel Rpc/vRPC.
- External traffic must pass through Portal entry/site/admission and then be delivered to the local App through Link webproxy.
- A direct App Web endpoint is an internal path that expects trusted headers written by Vine. Do not expose it publicly.
- Use Rpc/vRPC for structured health, readiness, and business probe results as well. Do not create extra JSON Web routes for them.

For the same-origin `:7288` RPCGW + WEBGW Standalone pattern, startup, production assets, and `webgw endpoint is unavailable` diagnosis, use [frontend-vrpc.md](frontend-vrpc.md#publish-a-local-standalone-browser-app-on-port-7288).

## Event and Task

### Consumer Semantics

| Question | Event | Task |
| --- | --- | --- |
| Meaning | A fact that already occurred | A request to perform work |
| Logical delivery | One copy for each distinct listening App name | One copy globally |
| Replica behavior | Instances with the same App name compete for that App's copy | All Runners compete for the same work |
| Sender waits for | Link to accept the publish | Link to accept the publish |
| Synchronous result | None | None |

Event does not broadcast to every process. Task is not directed back to the initiating App. When two logical consumers must both observe a fact, use Event Listeners with different App names. When horizontally scaling one logical consumer, share the same App name.

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

Defaults and limits in `v0.13.1`:

| Behavior | Event Listener | Task Runner |
| --- | --- | --- |
| Per-attempt timeout | 30 seconds | 30 seconds |
| Concurrency per registration instance | 10 | 10 |
| Failure | Retry | Retry |
| Retry limit | Not limited by Vine | Not limited by Vine |
| `NoRetry` | Terminal acknowledgement with no failure record | Terminal acknowledgement with no failure record |
| DLQ | None | None |
| Stream storage | Embedded NATS: memory; external NATS: deployment-owned | Embedded NATS: memory; external NATS: deployment-owned |

These are `v0.13.1` version facts (see [Version baseline quick reference](foundations.md#version-baseline-quick-reference) for the full list). Check the corresponding documentation and source when the target version differs.

### At-Least-Once Delivery and Cancellation

- Put a stable `eventId` or `jobId` in the contract. Protect side effects with unique constraints, transactional deduplication, idempotent upserts, or external idempotency keys.
- Lost acknowledgements, errors, timeouts, reconnects, or consumer reassignments can all cause duplicate execution.
- `NoRetry` only exchanges duplicate risk for terminal loss after the first failure. It does not create a DLQ or a queryable failure record.
- A timeout cancels one attempt's context but cannot stop a handler that ignores cancellation. An old attempt may overlap a retry.
- Do not rely on global ordering, replay for late Listeners, audit history, or exactly-once delivery.
- Embedded NATS uses memory streams. External NATS can use file-backed streams, but restart durability must be provisioned and recovery-tested by the deployment. If the business cannot tolerate loss, keep a database outbox, durable workflow, or another durable business record even when NATS is file-backed.

Event/Task propagates trace, sender App, and publication time. It does not propagate Actor, Initiator, the sender's deadline, or cancellation. For asynchronous authorization, include the minimum immutable identity in the contract and have the receiver revalidate it. Do not include credentials or secrets.

### Cron

`WithRunnerCronScheduler`:

- Accepts only a trigger with no input.
- Uses the trigger's Skel name, not the generated Go method name.
- Uses standard five-field cron syntax.
- Uses the Hub's clock and default local timezone.
- Does not promise to catch up executions missed while the runtime was unavailable.
- Enqueues scheduled messages in the same global Task queue with the same retry and idempotency semantics.

When strict timezones, calendars, misfire policies, execution history, or manual replay are required, use a dedicated scheduler or workflow system and have it launch Vine Tasks.

## vRPC over HTTP

Use Portal's vRPC over HTTP when browsers, mobile applications, external clients, or gateways access structured synchronous business APIs. Ordinary Go applications should prefer generated clients. Handle this wire protocol directly only when implementing a non-Go client, integrating a gateway, or troubleshooting at the wire level.

Request:

```text
POST <prefix>/<service-skel-name>/<method-skel-name>
```

Wire details in `v0.13.1`:

- The media type must be `application/vrpc+json` or `application/vrpc+cbor`, not ordinary `application/json`.
- Required request headers: `accept`, `content-type`, `vrpc-trace`, and `vrpc-client`.
- Optional headers: `vrpc-options`, `vrpc-actor`, and `vrpc-initiator`. External entries do not trust caller-forged Actor or Initiator values.
- JSON body shape: `{"params": {...}}`.
- Response body: `{"result": ..., "error": ...}`. Interpret the logical outcome using `vrpc-status` and `error.code`.
- The internal App handler's HTTP transport may always return 200 while storing logical status in `vrpc-status`. Portal maps this to conventional external 4xx/5xx statuses.
- Direct transport requires a trace ID and span. Portal may accept an ID-only trace and create the entry span.
- Portal defaults to 30 seconds, rejects timeouts above 120 seconds, clears external error details, and applies admission and CORS.

Do not rely only on the HTTP reason text, and do not write repeated lines for the same protocol header. Wire headers, media types, and JSON/CBOR fields are cross-component protocol. When changing them, update every producer, consumer, test, and compatibility note together.

## Validate Capabilities

- Use generated clients to validate Rpc success, business errors, SystemError, timeouts, and context propagation.
- Through Portal, validate admission, Actor, trace, and status mapping for every public Web/Rpc entry.
- Force Event/Task errors, timeouts, duplicate delivery, and cancellation, then verify idempotent state.
- Validate real consumer-group semantics across multiple App names and multiple replicas.
- Use standalone/testkit for business behavior and in-process value isolation. Use `vine dev` or separated processes for JSON/CBOR wire behavior, and separated processes for NATS, Link, leases, and networking.
- Pin request/response golden tests for wire integrations, including JSON and CBOR, header validation, and error mapping.

See [execution-and-boundaries.md](execution-and-boundaries.md) for context and error rules, and [official-sources.md](official-sources.md) for official sources.
