# Distributed Order System

A Go-based microservices project demonstrating distributed systems patterns
using RabbitMQ, PostgreSQL, and Redis.

## System Architecture

```
┌──────────────┐
│  API Gateway  │ :8080
│   (Go/Gin)    │
└──────┬───────┘
       │ HTTP
       ├──────────────────┬──────────────────┐
       ▼                  ▼                  ▼
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│    Order      │  │  Inventory   │  │   Payment    │
│   Service     │  │   Service    │  │   Service    │
│   :8002       │  │   :8001      │  │   :8003      │
└──────┬───────┘  └──────┬───────┘  └──────┬───────┘
       │                 │                 │
       │    PostgreSQL   │    RabbitMQ     │    Redis
       │  ┌──────────┐   │  ┌──────────┐   │  ┌───────┐
       └──┤  orders  │   └──┤ exchanges│   └──┤ claim │
          │  products│      │ queues   │      │ check │
          └──────────┘      └──────────┘      └───────┘
```

## Technologies

| Technology | Purpose |
|---|---|
| Go 1.26 | All services |
| Gin | HTTP framework |
| GORM | PostgreSQL ORM |
| PostgreSQL | Primary database per service |
| RabbitMQ | Asynchronous messaging (AMQP) |
| Redis | Claim Check payload storage |

## Service Overview

| Service | Port | Database Tables | Role |
|---|---|---|---|
| API Gateway | :8080 | None | HTTP proxy and webhook receiver |
| Order Service | :8002 | orders, order_products | Order lifecycle management |
| Inventory Service | :8001 | inventories, inventory_reservations | Stock tracking and reservation |
| Payment Service | :8003 | payments | Payment processing |

## Patterns Implemented

### CQRS (Command Query Responsibility Segregation)

Each service separates write operations (Commands) from read operations
(Queries) using distinct handler types.

```
┌─────────────────────────────────────────┐
│              Application Layer          │
│                                         │
│  ┌─────────────┐   ┌────────────────┐  │
│  │  Commands    │   │    Queries     │  │
│  │             │   │               │  │
│  │ CreateOrder │   │  GetOrder     │  │
│  │ FailOrder   │   │  ListOrders   │  │
│  │ UpdateStatus│   │               │  │
│  └──────┬──────┘   └───────┬───────┘  │
│         │                  │          │
│         ▼                  ▼          │
│  ┌─────────────┐   ┌────────────────┐  │
│  │ Repository  │   │  ReadModel     │  │
│  │ (write DB)  │   │  (read DB)    │  │
│  └─────────────┘   └────────────────┘  │
└─────────────────────────────────────────┘
```

**Implementation:**

- `services/order/internal/app/command/` - write handlers
- `services/order/internal/app/query/` - read handlers
- Domain entities enforce business rules on the write side
- Read models return flat DTOs optimized for queries

### Choreography Pattern (Event-Driven Saga)

Services coordinate through domain events without a central orchestrator.
Each service publishes events that others react to.

```
┌────────┐  OrderCreated   ┌────────────┐  StockReserved  ┌─────────┐
│  Order  │ ──────────────▶ │ Inventory  │ ──────────────▶ │ Payment │
│Service  │                 │  Service   │                 │ Service │
└────────┘                  └────────────┘                  └────────┘
     ▲                           ▲                              │
     │  PaymentSucceeded         │  PaymentSucceeded            │
     └───────────────────────────┴──────────────────────────────┘
```

**Compensation flows:**

```
StockRejected:  Inventory ──▶ Order (failed)
PaymentFailed:  Payment ──▶ Order (cancelled) ──▶ Inventory (restore stock)
```

### Competing Consumers

The Inventory Service runs multiple goroutines that compete for messages
from the same queue, enabling parallel processing with concurrency safety.

```
                    ┌──────────────┐
  inventories.orders│   RabbitMQ   │
  ─────────────────▶│    Queue     │
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
         ┌─────────┐ ┌─────────┐ ┌─────────┐
         │Worker 1 │ │Worker 2 │ │Worker 3 │
         └────┬────┘ └────┬────┘ └────┬────┘
              │            │            │
              ▼            ▼            ▼
         SELECT FOR UPDATE (row-level locks)
              │            │            │
              └────────────┼────────────┘
                           ▼
                      PostgreSQL
```

**Implementation:**

- 3 goroutines consume from `inventories.orders` queue
- `SELECT FOR UPDATE` acquires row-level locks within a transaction
- Manual acknowledgment ensures at-least-once delivery

### Compensating Transaction

Each step in the saga has a corresponding compensation that reverses its
effects when something fails downstream.

```
┌───────────┐                          ┌──────────────┐
│   Order    │─── publish fails ──────▶│  FailOrder   │
│  Created   │                          │  (mark failed)│
└───────────┘                          └──────────────┘

┌───────────┐                          ┌──────────────┐
│   Stock    │─── insufficient ───────▶│  StockRejected│
│  Reserved  │                          │  (order fails)│
└───────────┘                          └──────────────┘

┌───────────┐                          ┌──────────────┐
│  Payment   │─── payment fails ──────▶│ Cancel Order │───▶ Restore Stock
│  Processed │                          │ Cancel Resv  │
└───────────┘                          └──────────────┘
```

**Implementation:**

- `services/order/internal/app/command/create_order.go` - publish failure
  triggers `FailOrder`
- `services/inventory/internal/app/command/cancel_reservation.go` - restores
  stock on payment failure
- `services/order/internal/app/command/fail_order.go` - marks orders as failed

### Retry (Message Processing)

All consumers implement a retry queue with dead-letter exchange pattern.
Failed messages are retried up to 5 times before moving to a DLQ.

```
┌───────────┐     fail      ┌────────────┐   TTL expires   ┌───────────┐
│ Main Queue │ ────────────▶ │Retry Queue │ ──────────────▶ │Main Queue │
│            │               │ (5s TTL)   │                 │  (retry)  │
└─────┬─────┘               └────────────┘                 └───────────┘
      │                            │
      │  max retries exceeded      │
      ▼                            ▼
┌───────────┐               ┌───────────┐
│    DLQ    │               │x-retry-count│
│           │               │ incremented │
└───────────┘               └───────────┘
```

**Per-service retry infrastructure:**

| Service | Queue | Retry Queue | DLQ | Max Retries |
|---|---|---|---|---|
| Inventory (orders) | inventories.orders | inventories.orders.retry | inventories.orders.dlq | 5 |
| Inventory (payments) | inventories.payments | inventories.payments.retry | inventories.payments.dlq | 5 |
| Order (inventory) | orders.inventories | orders.inventories.retry | orders.inventories.dlq | 5 |
| Order (payments) | orders.payments | orders.payments.retry | orders.payments.dlq | 5 |
| Payment (inventory) | payments.inventories | payments.inventories.retry | payments.inventories.dlq | 5 |

### Claim Check Pattern

Large order payloads are stored in Redis while only a lightweight reference
is sent through RabbitMQ, keeping the message broker efficient.

```
┌──────────────┐                     ┌──────────────┐
│    Order      │                     │  Inventory   │
│   Service     │                     │   Service    │
│  (Publisher)  │                     │  (Consumer)  │
└──────┬───────┘                     └──────┬───────┘
       │                                    │
       │ 1. Store full payload in Redis     │
       │    key: claim-check:orders:{id}    │
       │    TTL: 1 hour                     │
       │                                    │
       │ 2. Send lightweight message        │
       │    {orderId, claimCheckKey}        │
       │ ──────────────────────────────────▶│
       │         RabbitMQ                   │
       │                                    │
       │                                    │ 3. Fetch payload
       │                                    │    from Redis
       │                                    │
       │                                    │ 4. Remove claim check
       │                                    │    from Redis
```

## Message Flow

### Happy Path - Checkout and Payment

```
User ──POST /api/orders──▶ API Gateway ──▶ Order Service
                                                  │
                                          Save order (pending)
                                          Store payload in Redis
                                          Publish OrderCreated
                                                  │
                                          ┌───────▼────────┐
                                          │ Inventory Svc  │
                                          │ (3 workers)    │
                                          │                │
                                          │ Fetch from Redis│
                                          │ Reserve stock  │
                                          │ Publish        │
                                          │ StockReserved  │
                                          └───────┬────────┘
                                                  │
                                          ┌───────▼────────┐
                                          │ Payment Svc    │
                                          │                │
                                          │ Create payment │
                                          │ (pending)      │
                                          │ Trigger webhook│
                                          └───────┬────────┘
                                                  │
User ──POST /api/payments/{id}/process──▶ Payment Service
                                                  │
                                          Publish PaymentSucceeded
                                                  │
                                     ┌────────────┼────────────┐
                                     ▼                         ▼
                              Order Service           Inventory Service
                              (confirmed)             (reservation completed)
```

### Unhappy Path - Stock Rejected

```
Order Service ──OrderCreated──▶ Inventory Service
                                      │
                                Insufficient stock
                                      │
                                Publish StockRejected
                                      │
                                Order Service (failed)
```

### Unhappy Path - Payment Failed

```
Payment Service ──PaymentFailed──▶ Order Service (cancelled)
                                      │
                                 Inventory Service
                                 (cancel reservation, restore stock)
```

## Architecture

Each service follows **hexagonal architecture** (ports and adapters):

```
┌─────────────────────────────────────────────┐
│                  Adapters                    │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐  │
│  │   HTTP   │  │   DB     │  │ Messaging  │  │
│  │ Handlers │  │Repository│  │Pub / Sub   │  │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘  │
│       │             │              │         │
│  ─────┴─────────────┴──────────────┴───────  │
│                  Ports (interfaces)           │
│  ─────────────────────────────────────────── │
│                                             │
│              Application Layer              │
│        (Commands, Queries, Handlers)        │
│                                             │
│  ─────────────────────────────────────────── │
│                                             │
│               Domain Layer                  │
│      (Entities, Value Objects, Rules)       │
│                                             │
└─────────────────────────────────────────────┘
```

- **Domain layer** - pure Go with no external dependencies
- **Application layer** - command/query handlers with dependency injection
- **Adapter layer** - HTTP handlers, DB repositories, message publishers/consumers

## Key Concepts

- **Eventual consistency** - services converge to a consistent state via events
- **Message deduplication** - idempotent handlers prevent duplicate processing
- **Idempotency** - domain state machines reject invalid state transitions
