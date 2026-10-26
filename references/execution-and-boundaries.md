# Vine Execution, Context, and Error Boundaries

This reference is based on Vine `v0.32.5`. Use it for DI, execution, Filter, context, Actor, trace, timeout, and cross-boundary error tasks. Resolve exact symbols against public GoDoc for the target project's pinned version. The timeout values here are `v0.32.5` version facts; the authoritative list is in [Version baseline quick reference](foundations.md#version-baseline-quick-reference).

## Contents

- [Two object lifecycles](#two-object-lifecycles)
- [DI binding and scope](#di-binding-and-scope)
- [Execution and Filter](#execution-and-filter)
- [Context, Trace, and identity](#context-trace-and-identity)
- [Timeout and cancellation](#timeout-and-cancellation)
- [Structured errors](#structured-errors)
- [Review checklist](#review-checklist)

## Two Object Lifecycles

### App-Lifecycle Singletons

Components and Modules are each constructed once during `Start()` and retained until the App stops. They receive the application root `context.Context` and root `meta.Context`.

A generated Rpc client injected into a Module represents a background call initiated by the current App. Its lifetime and metadata remain tied to the App root. Resolve request-aware objects within the Handler or Filter execution and release them with that execution.

### Execution Lifecycle

Every Rpc, Web, Event, or Task delivery creates an independent execution:

```text
protocol entry → seed context/metadata → resolve filters → resolve handler → invoke target
               → filters return outward → dispose execution-owned objects in reverse order
```

Handlers, Listeners, Runners, and Filters are constructed inside the execution. The execution is complete after the Filter chain returns. Complete dependency resolution and request-object use before the execution ends.

## DI Binding and Scope

Typical bindings:

```go
b.Bind(di.T[Gateway]()).
    ToImplementation(di.T[*SMTPGateway]()).
    In(di.SingletonScope)

b.BindFactory(func(ctx context.Context) *Repository {
    return NewRepository(ctx)
})

b.BindInstance(existing)
```

Use `inject:""` on exported fields. A Factory may return `T` or `(T, error)`; a non-nil error becomes a panic.

Critical fallback rules:

| Binding | Plain/root injector | Execution injector |
| --- | --- | --- |
| Default scope | New transient on every resolution | One instance per execution |
| `SingletonScope` | Root-container singleton | Reuse the root-container singleton |
| `ExecutionScope` | Requires an execution injector | One instance per execution |
| `TransientScope` | New instance on every resolution | New instance on every resolution |

Scope attaches to the requested target type; the concrete implementation has its own binding and scope. When an interface and implementation must share the same instance, keep their scopes explicitly aligned or forward both to the same instance or factory.

Rules:

- Give singleton dependencies application-compatible lifetimes; resolve execution-scoped dependencies within an execution.
- Use default scope for context-aware clients, configuration, DAOs, caches and lockers. Select an explicit scope when its lifecycle requirements are established.
- Resolve request clients and dynamic configuration readers in the current execution so they capture its context and configuration.
- `DIInit()` runs after field injection.
- Execution-owned objects can be released in reverse order through `DIDispose()` or a disposer.
- The protocol entry owns and releases seeded protocol objects.
- Close singleton databases, Redis, files, queue connections and goroutines through their owning Component or Module lifecycle.

`BindCommon`, Component `Bind`, Module `Bind`, and capability `...Bind` methods may be applied to multiple containers. Keep binding declarations deterministic and repeatable; perform I/O, resource startup and registration in lifecycle hooks.

## Execution and Filter

Filter implementation:

```go
type TimingFilter struct {
    Context *ctr.Context `inject:""`
}

func (f *TimingFilter) Filter(next ctr.FilterNext) {
    started := time.Now()
    next()
    _ = time.Since(started)
}
```

Filters form an onion in declaration order:

```text
FilterA before
  FilterB before
    target
  FilterB after
FilterA after
```

A Filter may read or rewrite the target, method and arguments before invocation. After invocation, it may inspect the completed call and rewrite results.

A short-circuiting Filter completes Rpc, Event or Task execution by setting results compatible with the target contract. The executor validates that result shape.

Use Filters for authentication, logging, metrics, and protocol adaptation shared across Handlers. Keep business state transitions in business services or Handlers.

## Context, Trace, and Identity

`meta.Context` extends `context.Context` with:

- `Trace`
- `Initiator`
- `Actor`

Distinguish Actor states:

| State | Meaning |
| --- | --- |
| absent | Identity metadata is absent |
| anonymous | A known unauthenticated external identity |
| authenticated | Generated Actor info established by entry policy |

Use `meta.GetActorInfo[T]` to read registered generated Actor info. Preserve the distinction between absent identity and an explicitly anonymous Actor.

Trust boundaries:

- Portal applies entry, authentication, and authorization policy to external requests, then writes backend Actor and Initiator metadata.
- Portal derives trusted `vrpc-actor`, `vrpc-initiator`, `vweb-actor` and `vweb-initiator` values from admission policy and replaces caller-supplied values.
- Restrict direct App and Link endpoint access to trusted internal callers.

A trace ID is 32 lowercase hexadecimal characters and a span is 16; each must contain at least one nonzero digit. When creating a child trace, preserve the trace ID, use the current span as the local parent, and generate a new span. `ParentSpan` records the local parent span. Complete OpenTelemetry span modeling requires a separate tracing integration.

A generated Rpc client injected into an inbound execution inherits trace, Actor, Initiator, and the remaining deadline. Background calls from a Module use the App root context, usually creating a new trace with an absent Actor.

Derive local child contexts from the injected request context to preserve cancellation, deadlines and metadata lifecycle.

Event/Task propagates trace, sender App and publication time. Each receiver starts with absent Actor/Initiator metadata and its own deadline and cancellation. Put the minimum immutable identity needed for asynchronous authorization into the contract and have the receiver revalidate it. Keep credentials and secrets in the owning authentication system.

## Timeout and Cancellation

In `v0.32.5`:

- Rpc defaults to 30 seconds; Portal rejects Rpc/MCP requested budgets above 120 seconds.
- The ordinary Web fallback policy defaults to 30 seconds and accepts requests up to 120 seconds. A site `webTimeout.default` or most-specific `paths` rule selects the total budget; the request header may select a shorter budget, bounded by that policy. Paths match the decoded, rewritten backend path. Configured positive durations can exceed the legacy request maximum.
- Budget begins at the Portal entry and includes admission, authentication, and checks. Each forward carries the remaining time.
- A generated downstream client inherits the remaining deadline when it retains the injected context.
- With the default stream policy, SSE/WebSocket lasts for the connection lifetime subject to a 60-second traffic-idle limit. An explicit total budget adds a deadline. Portal owns SSE idle enforcement; internal Link/App streams preserve deadline, cancellation and truncation errors.

These are `v0.32.5` version facts (see [Version baseline quick reference](foundations.md#version-baseline-quick-reference) for the full list). Check the corresponding documentation and source when the target version differs.

A timeout cancels the context; the handler must observe cancellation to stop its work. For Event/Task, an old attempt may continue after timeout and overlap a retry. Check context in long loops, before external I/O, and before irreversible side effects.

Ordinary Web, Rpc and MCP executions accepted by Portal continue after client disconnect, bounded by the gateway execution budget and shutdown. Streaming work with the default budget remains tied to its live connection. Transaction commit and rollback follow the application transaction policy.

## Structured Errors

Use stable `core/ex` `Code` values for errors crossing boundaries:

```go
err := ex.New(
    ex.InvalidRequest,
    "invalid checkout request",
    ex.WithReason("currency_not_supported"),
    ex.WithDetail(detail),
    ex.WithCause(cause),
)
```

Rules:

- Consumers classify errors by `Code` and agreed `Reason` values.
- `WithCause` supports in-process `errors.Is`; cross-process classification uses serialized Code and Reason.
- Use registered `Code` values; invalid or unregistered values raise a panic.
- Use `Internal` and `Unknown` for boundary fallbacks and registered business Codes for expected business failures.
- `Recover` catches any `ex.Error`. `RecoverApplication` catches only business errors and continues to panic for SystemError and non-Vine panics.
- Portal preserves the external detail field shape and clears its contents. Put client-facing classification in Code and Reason.

Low-level Rpc clients panic on SystemError by default. Use `ReturnIfSystemError` or a generated ER client only when explicit error returns are truly required, and preserve the same error classification semantics.

## Review Checklist

- [ ] Execution objects are used and released within their owning execution.
- [ ] Child calls derive from the injected context and preserve the active call chain.
- [ ] Bind methods declare deterministic bindings; lifecycle hooks own I/O, startup and registration.
- [ ] Filter short-circuits produce valid results, and argument/result rewrites match the contract.
- [ ] Actor comes only from a trusted entry or an explicit internal execution.
- [ ] Timeout and retry operations are idempotent and observe cancellation.
- [ ] Errors use registered Codes, and public messages/details contain sanitized client-facing information.
- [ ] Resources are released by their real owner through an execution disposer or App lifecycle.

See [official-sources.md](official-sources.md) for official pages.
