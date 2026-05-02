# Message Topology - RabbitMQ Exchanges and Queues

```mermaid
graph LR
    subgraph Exchanges
        OE["orders<br/>(topic)"]
        IE["inventories<br/>(topic)"]
        PE["payments<br/>(topic)"]
    end

    subgraph OrderQueues ["Order Service Queues"]
        OI["orders.inventories"]
        OP["orders.payments"]
    end

    subgraph InventoryQueues ["Inventory Service Queues"]
        IO["inventories.orders"]
        IP["inventories.payments"]
    end

    subgraph PaymentQueues ["Payment Service Queues"]
        PI["payments.inventories"]
    end

    subgraph Publishers
        OrderPub["Order Service"]
        InvPub["Inventory Service"]
        PayPub["Payment Service"]
    end

    OrderPub -->|"orders.created"| OE
    InvPub -->|"inventories.reserved<br/>inventories.rejected"| IE
    PayPub -->|"payments.succeeded<br/>payments.failed"| PE

    OE -->|"orders.created"| IO
    IE -->|"inventories.reserved"| PI
    IE -->|"inventories.rejected"| OI
    PE -->|"payments.succeeded<br/>payments.failed"| OP
    PE -->|"payments.succeeded<br/>payments.failed"| IP
```

## Exchange Details

| Exchange | Type | DLX | Declared By |
|---|---|---|---|
| `orders` | topic | `orders.dlx` | Order Service publisher |
| `inventories` | topic | `inventories.dlx` | Inventory Service publisher |
| `payments` | topic | `payments.dlx` | Payment Service publisher |

## Queue Bindings

| Queue | Exchange | Routing Key | Consumer |
|---|---|---|---|
| `inventories.orders` | `orders` | `orders.created` | Inventory (3 workers) |
| `orders.inventories` | `inventories` | `inventories.rejected` | Order |
| `orders.payments` | `payments` | `payments.*` | Order |
| `payments.inventories` | `inventories` | `inventories.reserved` | Payment |
| `inventories.payments` | `payments` | `payments.*` | Inventory |
