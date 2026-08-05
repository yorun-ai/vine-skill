# Vine Execution, Context, and Error Boundaries

This reference is based on Vine `v0.12.0`. Use it for DI, execution, Filter, context, Actor, trace, timeout, and cross-boundary error tasks. Resolve exact symbols against public GoDoc for the target project's pinned version.

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

Components and Modules are each constructed once during `Start()` and retained until the App stops. They receive the application root `context.Context` and root `meta.Context`, not the context of a future request.

A generated Rpc client injected into a Module represents a background call initiated by the current App. It does not later become request-aware automatically. Do not retain execution objects obtained in Handlers or filters inside Modules, Components, global variables, caches, or background goroutines.

### Execution Lifecycle

Every Rpc, Web, Event, or Task delivery creates an independent execution:

```text
protocol entry → seed context/metadata → resolve filters → resolve handler → invoke target
               → filters return outward → dispose execution-owned objects in reverse order
```

Handlers, Listeners, Runners, and Filters are constructed inside the execution. The execution is complete after the Filter chain returns. Do not continue resolving dependencies or using canceled request objects afterward.

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
| No explicit scope | New transient on every resolution | One instance per execution |
| `SingletonScope` | Root-container singleton | Reuse the root-container singleton |
| `ExecutionScope` | Cannot resolve | One instance per execution |
| `TransientScope` | New instance on every resolution | New instance on every resolution |

Scope attaches to the requested target type; it does not automatically propagate to the concrete implementation. When an interface and implementation must share the same instance, keep their scopes explicitly aligned or forward both to the same instance or factory.

Rules:

- Do not let a singleton depend on an explicitly execution-scoped type.
- Keep context-aware clients, configuration, DAOs, caches, and lockers unscoped unless a stronger lifecycle requirement has been proven.
- Do not make a request client or dynamic configuration reader a singleton merely to avoid construction. Doing so freezes the context or configuration from the first resolution.
- `DIInit()` runs after field injection.
- Execution-owned objects can be released in reverse order through `DIDispose()` or a disposer.
- Seeded protocol objects are not owned by DI and are not disposed by DI.
- A plain injector does not automatically close singletons. Close databases, Redis, files, queue connections, and goroutines through their owning Component or Module lifecycle.

`BindCommon`, Component `Bind`, Module `Bind`, and capability `...Bind` methods may be applied to multiple containers. Make them declare bindings only; keep them deterministic, repeatable, and free of runtime side effects.

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

A Filter may read or rewrite the target, method, and arguments before invocation, and may rewrite results afterward. It cannot change the target, method, or arguments after target execution has completed.

Short-circuiting is not just omitting `next()`: Rpc, Event, and Task executors still require a valid result shape. A short-circuiting Filter must set results compatible with the target contract, or the execution fails because results are missing.

Use Filters for authentication, logging, metrics, and protocol adaptation shared across Handlers. Keep business state transitions in business services or Handlers so that Filters do not become hidden business workflows.

## Context, Trace, and Identity

`meta.Context` extends `context.Context` with:

- `Trace`
- `Initiator`
- `Actor`

Distinguish Actor states:

| State | Meaning |
| --- | --- |
| absent | No identity was provided |
| anonymous | A known unauthenticated external identity |
| authenticated | Generated Actor info established by entry policy |

Use `meta.GetActorInfo[T]` to read registered generated Actor info. Do not conflate absent and anonymous.

Trust boundaries:

- Portal applies entry, authentication, and authorization policy to external requests, then writes backend Actor and Initiator metadata.
- External callers cannot gain a trusted identity through self-supplied `vrpc-actor`, `vrpc-initiator`, `vweb-actor`, or `vweb-initiator` values.
- Direct App and Link endpoints are internal boundaries and should not be exposed to external clients.

A trace ID is 32 lowercase hexadecimal characters and a span is 16; neither may be all zeros. When creating a child trace, preserve the trace ID, use the current span as the local parent, and generate a new span. `ParentSpan` is local information, not a complete OpenTelemetry span model.

A generated Rpc client injected into an inbound execution inherits trace, Actor, Initiator, and the remaining deadline. Background calls from a Module use the App root context, usually creating a new trace with an absent Actor.

Never replace an active request context with `context.Background()`. Doing so loses cancellation, deadlines, and metadata lifecycle. Derive a child from the injected context when a local child is needed.

Event/Task propagates only trace, sender App, and publication time. It does not propagate Actor, Initiator, the sender's deadline, or cancellation. Put the minimum immutable identity needed for asynchronous authorization into the contract and have the receiver revalidate it. Do not copy credentials or secrets.

## Timeout and Cancellation

In `v0.12.0`:

- Ordinary Rpc and Web have a default total timeout of 30 seconds.
- Portal rejects request timeouts above 120 seconds.
- Budget begins at the Portal entry and includes admission, authentication, and checks. Each forward carries the remaining time.
- A generated downstream client inherits the remaining deadline when it retains the injected context.
- SSE/WebSocket has no total timeout when none is explicitly set, but uses a 60-second traffic-idle timeout.

These are version-specific facts. Check the corresponding documentation and source when the target version differs.

A timeout cancels the context but cannot force Go code that ignores cancellation to stop. For Event/Task, an old attempt may continue after timeout and overlap a retry. Check context in long loops, before external I/O, and before irreversible side effects.

An ordinary external client disconnect does not necessarily cancel a backend execution already accepted by Portal. Do not treat a network disconnect as an automatic business transaction rollback signal.

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

- Consumers should inspect `Code` and agreed `Reason` values, not parse messages or depend on concrete types.
- `WithCause` supports in-process `errors.Is`, but the cause is not serialized.
- An unregistered or invalid `Code` panics. Do not assemble unknown error codes by hand.
- `Internal` and `Unknown` are suitable boundary fallbacks, not generic business errors.
- `Recover` catches any `ex.Error`. `RecoverApplication` catches only business errors and continues to panic for SystemError and non-Vine panics.
- Portal preserves the external detail field shape but clears its contents. Do not rely on internal details leaking to clients.

Low-level Rpc clients panic on SystemError by default. Use `ReturnIfSystemError` or a generated ER client only when explicit error returns are truly required, and preserve the same error classification semantics.

## Review Checklist

- [ ] No execution object escapes into a singleton, cache, or goroutine.
- [ ] No active call chain is severed with `context.Background()`.
- [ ] Bind methods perform no I/O, startup, registration, or non-repeatable side effects.
- [ ] Filter short-circuits produce valid results, and argument/result rewrites match the contract.
- [ ] Actor comes only from a trusted entry or an explicit internal execution.
- [ ] Timeout and retry operations are idempotent and observe cancellation.
- [ ] Errors use registered Codes, and public messages/details reveal no internal or sensitive information.
- [ ] Resources are released by their real owner through an execution disposer or App lifecycle.

See [official-sources.md](official-sources.md) for official pages.
