# Two Domains: Orders and Catalog

This example uses Vine `v0.32.5`, skelc `v0.34.0` and Go `1.27.3`. It demonstrates two domains consuming each other's generated public contracts: synchronous Rpc in both directions and an Event round trip. Read [foundations.md](../foundations.md) for the baseline and [capabilities.md](../capabilities.md) for execution and delivery semantics.

## Domain and App Boundaries

| Boundary | Responsibility |
| --- | --- |
| `shop.catalog` domain | Product DTO, product lookup service and reservation-created Event |
| `shop.orders` domain | Order DTO, order lookup service, order-placed Event and browser API |
| `shop.catalog` App | Implements catalog handlers and consumes order Events |
| `shop.orders` App | Implements order handlers and consumes catalog Events |

A domain names and owns contracts. An App owns handler registration, DI and lifecycle. This example gives each domain its own App; the matching names are an explicit choice. Multiple domains can be assembled into one App when that fits the application. Event consumer groups use the listening App name; distinct domain names alone do not create separate logical deliveries.

The launcher bundles both Apps into one process with a shared Hub, Link, Portal and embedded NATS. Each App has its own registrations and state. Cross-App calls use generated Rpc clients, and Events pass through Link/NATS. See [app.go](src/server/app/app.go) for `standalone.NewBundledWithOption` assembly.

## Complete Flow

```text
Portal → shop.orders.OrderApiService.placeOrder(sku, quantity)
  orders → catalog.CatalogService.getProduct                    [Rpc]
  orders stores the order
  orders → OrderPlacedEvent → catalog.OrderPlacedListener       [Event]
  catalog → orders.OrderService.getOrder                        [Rpc]
  catalog records the reservation
  catalog → ReservationCreatedEvent → orders listener           [Event]
Portal → shop.orders.OrderApiService.getOrder → confirmed order
```

The reverse Rpc runs in the catalog Listener's execution. Its result supplies the product title and quantity for the confirmation. The final Listener updates the order; the workflow ends there. The placement response confirms storage and accepted publication, and callers query order status for asynchronous completion.

## Contracts and Generation

Start with the [catalog contract](skel/catalog/catalog.skel) and [orders contract](skel/orders/orders.skel). `pub service` exposes a backend interface; `api service` exposes the separate Portal entry. `pub event` lets other domains subscribe to the owning domain's facts.

The orders domain references the catalog's public DTO:

```skel
domain shop.orders
import shop.catalog as catalog

pub data Order {
    id: uuid
    product: catalog.Product
    quantity: int
    confirmed: bool
    confirmation: string
}
```

The canonical [orders domain header](skel/orders/domain.skel) declares the import. Generation needs two mappings:

- `--skel-import shop.catalog=./skeled/skel/catalog` resolves the published Skel contract.
- `--go-import shop.catalog=example.com/shop/skeled/catalogpub@v0.0.0` selects the generated Go package and its module version.

These mappings have separate roles: compiler input resolution and generated language dependency resolution. `skelc check` checks local syntax/semantics with unresolved imports; generation resolves and validates the imported contracts.

[generate.sh](generate.sh) first exports catalog's public Skel and generates its Go modules, then resolves them while generating orders. Each domain uses `--go-out` and `--go-pub-out` together:

| Declaration | Owner's regular package | Consumer's public package |
| --- | --- | --- |
| `pub service` | Server, default Server and service registration | Client and ER Client |
| `pub event` | Emitter and publication registration | Listener, default Listener and subscription registration |
| Public DTO | Facade over the public type | Canonical shared type |

For example, orders imports `catalogpub.CatalogServiceClient` and `catalogpub.DefaultReservationCreatedEventListener`; catalog imports `orderspub.OrderServiceClient` and `orderspub.DefaultOrderPlacedEventListener`. Each side imports its own regular package for Servers and Emitters. Use the paired generation outputs so client/server and emitter/listener specs register on the correct side.

The public schema dependency is one-way: `orders.Order` contains `catalog.Product`. The catalog contract uses its own scalars and types; its Go Listener can consume `orderspub` directly without adding a Skel type dependency. **Schema dependencies follow referenced types; runtime call direction is independent.** Keep contract dependencies acyclic by choosing ownership of shared DTOs. A Skel import is needed when a declaration references another domain's types; a Go import supplies the generated client or Listener used by hand-written code.

`ext service` and `ext event` define extension points with reversed ownership: consumers implement the exported Server or use the exported Emitter. This example uses ordinary `pub` ownership throughout.

## Read the Implementation

| Source | What to follow |
| --- | --- |
| [orders/order_service.go](src/server/orders/order_service.go) | Injects catalog's generated Client, calls `GetProduct`, stores an order and emits `OrderPlaced`; also implements the public order lookup Server. |
| [catalog/catalog_service.go](src/server/catalog/catalog_service.go) | Implements catalog's own generated Server. |
| [catalog/order_placed_listener.go](src/server/catalog/order_placed_listener.go) | Embeds orders' public default Listener, calls the generated `OrderServiceClient`, and emits catalog's confirmation. |
| [orders/reservation_created_listener.go](src/server/orders/reservation_created_listener.go) | Embeds catalog's public default Listener and applies the confirmation. |
| [catalog/app.go](src/server/catalog/app.go) and [orders/app.go](src/server/orders/app.go) | Enable and register Services and Listeners in their respective Apps. |

Generated Clients and Emitters are injected within each request/Event execution. Vine supplies its context, trace, deadline and Link transport. App-owned Modules hold only business state: [OrderStore](src/server/orders/order_store.go) and [Reservations](src/server/catalog/reservations.go). The two domain implementation packages communicate through public contracts.

## Build and Try

Copy this directory into a disposable working directory before generating. Prerequisites: the pinned Go and skelc, Bash and access to Go modules.

```bash
bash generate.sh
cd src/server
GOWORK=off go test -race -v ./...
GOWORK=off go vet ./...
# After checking ports 7388 and 7399 are available:
GOWORK=off go run ./cmd/shop
```

- Business API prefix: `http://127.0.0.1:7388/api/invoke`.
- API: `shop.orders.OrderApiService.placeOrder` takes `{sku: "notebook", quantity: 2}`; `getOrder` takes `{orderId: "<returned UUID>"}`.
- Hub Dashboard/Admin: `http://127.0.0.1:7399/`, unauthenticated and bound to loopback.
- Stop the foreground process with Ctrl+C.

The package command `bash scripts/validate_example.sh` generates a disposable copy and runs the integration suite automatically. `generate.sh` keeps all generated output under `skeled/`; it supplies local `replace` directives for each module being built. Go does not inherit a dependency module's `replace` directives. The public Skel and Go outputs have separate directories because each generator owns its output directory.

## Validation and Scope

[domains_test.go](src/server/app/domains_test.go) starts both real Apps on an isolated Portal port and places concurrent orders. It checks the imported Product DTO returned by orders → catalog Rpc, confirmation content obtained by catalog → orders Rpc, both Event deliveries, and propagation of a catalog `NOT_FOUND` error through orders to Portal. Race checks exercise the shared module state.

Order/reservation state and deduplication are in memory for this example. One creation/reservation per order uses the order ID as its stable event identity; repeated catalog deliveries reuse the recorded notification and repeated confirmations preserve the final state. A restart clears both stores. Production recovery requires durable state, transactional outbox/inbox handling and a reviewed MQ storage policy. The [Greeting example](../example-greeting/README.md#event-to-task-receipts) separately demonstrates SQLite deduplication and injected delivery failures/retries.

This example verifies two App identities in one Standalone process. Cross-process transport, external NATS restart durability, multiple replicas, multi-consumer fan-out and failure recovery require tests in the corresponding deployment topology.
