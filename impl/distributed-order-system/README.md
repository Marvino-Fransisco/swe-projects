# Distributed Order System

A Go-based microservices project demonstrating distributed systems patterns
using RabbitMQ, PostgreSQL, and Redis.

## System Architecture

See [System Architecture Diagram](diagrams/system-architecture.md) for the full
Mermaid diagram.

## Diagrams

All system diagrams are in the [`diagrams/`](diagrams/) directory:

| Diagram | Description |
|---|---|
| [System Architecture](diagrams/system-architecture.md) | Services, infrastructure, and network topology |
| [Checkout Order](diagrams/checkout-order.md) | Sequence diagram for the checkout flow |
| [Process Payment](diagrams/process-payment.md) | Sequence diagram for payment processing |
| [Choreography Saga](diagrams/choreography-saga.md) | Event-driven saga and event catalog |
| [Message Topology](diagrams/message-topology.md) | RabbitMQ exchanges, queues, and bindings |
| [Hexagonal Architecture](diagrams/hexagonal-architecture.md) | Ports and adapters per service |
| [CQRS Pattern](diagrams/cqrs-pattern.md) | Command/Query separation per service |
| [Competing Consumers](diagrams/competing-consumers.md) | Inventory workers with row-level locking |
| [Compensating Transactions](diagrams/compensating-transaction.md) | Failure recovery and compensation flows |
| [Retry with DLQ](diagrams/retry-dlq.md) | Message retry and dead-letter queue pattern |
| [Claim Check](diagrams/claim-check.md) | Large payload optimization with Redis |

## Technologies

| Technology | Purpose |
|---|---|
| Go 1.26 | All services |
| Gin | HTTP framework |
| GORM | PostgreSQL ORM |
| PostgreSQL 18 | Primary database (shared instance, per-service tables) |
| RabbitMQ 4.3 | Asynchronous messaging (AMQP, topic exchanges) |
| Redis 8 | Claim Check payload storage |

## Service Overview

| Service | Port | Database Tables | HTTP Routes | Role |
|---|---|---|---|---|
| API Gateway | :8080 | None | `POST /api/orders`, `GET /api/inventories`, `POST /api/payments/:id/process`, `POST /api/webhooks` | HTTP proxy and webhook receiver |
| Inventory Service | :8001 | inventories, inventory_reservations | `GET /api/inventories` | Stock tracking and reservation |
| Order Service | :8002 | orders, order_products | `POST /api/orders`, `GET /api/orders`, `GET /api/orders/:id` | Order lifecycle management |
| Payment Service | :8003 | payments | `POST /api/payments/:payment_id/process` | Payment processing |

## Patterns Implemented

### Hexagonal Architecture (Ports and Adapters)

Each service follows hexagonal architecture with three layers:

See [Hexagonal Architecture Diagram](diagrams/hexagonal-architecture.md).

- **Domain layer** (`internal/domain/`) - pure Go with no external dependencies,
  contains entities, value objects, and business rules via state machines
- **Application layer** (`internal/app/`) - command/query handlers with
  dependency injection through port interfaces
- **Adapter layer** (`internal/adapters/` + `messaging/`) - HTTP handlers
  (Gin), DB repositories (GORM), message publishers and consumers (RabbitMQ)

### CQRS (Command Query Responsibility Segregation)

Each service separates write operations (Commands) from read operations (Queries)
using distinct handler types.

See [CQRS Pattern Diagram](diagrams/cqrs-pattern.md).

| Service | Commands | Queries |
|---|---|---|
| Order | CreateOrder, FailOrder, UpdateOrderStatus | GetOrder, ListOrders |
| Inventory | ReserveStock, CompleteReservation, CancelReservation | ListInventories |
| Payment | CreatePayment, ProcessPayment | None |

### Choreography Pattern (Event-Driven Saga)

Services coordinate through domain events without a central orchestrator.
Each service publishes events that others react to.

See [Choreography Saga Diagram](diagrams/choreography-saga.md).

### Competing Consumers

The Inventory Service runs 3 goroutines that compete for messages from the same
queue, enabling parallel processing with concurrency safety via `SELECT FOR UPDATE`.

See [Competing Consumers Diagram](diagrams/competing-consumers.md).

### Compensating Transaction

Each step in the saga has a corresponding compensation that reverses its effects
when something fails downstream.

See [Compensating Transactions Diagram](diagrams/compensating-transaction.md).

### Retry (Message Processing)

All consumers implement a retry queue with dead-letter exchange pattern.
Failed messages are retried up to 5 times with 5s TTL before moving to a DLQ.

See [Retry with DLQ Diagram](diagrams/retry-dlq.md).

### Claim Check Pattern

Large order payloads are stored in Redis while only a lightweight reference
(`claim-check:orders:{id}`, TTL 1 hour) is sent through RabbitMQ, keeping the
message broker efficient.

See [Claim Check Diagram](diagrams/claim-check.md).

## Message Topology

See [Message Topology Diagram](diagrams/message-topology.md) for the full
exchange and queue topology.

### Exchanges

| Exchange | Type | DLX |
|---|---|---|
| `orders` | topic | `orders.dlx` |
| `inventories` | topic | `inventories.dlx` |
| `payments` | topic | `payments.dlx` |

### Queue Bindings

| Queue | Exchange | Routing Key | Consumer |
|---|---|---|---|
| `inventories.orders` | `orders` | `orders.created` | Inventory (3 workers) |
| `orders.inventories` | `inventories` | `inventories.rejected` | Order |
| `orders.payments` | `payments` | `payments.*` | Order |
| `payments.inventories` | `inventories` | `inventories.reserved` | Payment |
| `inventories.payments` | `payments` | `payments.*` | Inventory |

## Key Concepts

- **Eventual consistency** - services converge to a consistent state via events
- **Message deduplication** - idempotent handlers prevent duplicate processing
- **Idempotency** - domain state machines reject invalid state transitions
- **Manual ACK** - consumers explicitly acknowledge after successful processing
- **Context-based transactions** - `shared/tx` package for transparent DB tx propagation
- **Dependency injection** - constructor-based DI via `bootstrap/router.go`
