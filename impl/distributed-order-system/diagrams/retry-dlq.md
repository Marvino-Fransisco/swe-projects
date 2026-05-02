# Retry with Dead-Letter Queue - Error Handling

```mermaid
graph LR
    subgraph QueueTopology ["Per-Consumer Queue Topology"]
        Main["Main Queue<br/>(e.g., inventories.orders)"]
        Retry["Retry Queue<br/>(5s TTL)"]
        DLQ["Dead-Letter Queue<br/>(DLQ)"]
        DLX["Main Exchange"]
        RetryDLX["Retry DLX"]
    end

    Main -->|"nack (requeue=false)"| RetryDLX
    RetryDLX --> Retry
    Retry -->|"TTL expires"| DLX
    DLX -->|"re-deliver"| Main

    Main -->|"max retries exceeded"| DLQ
    Retry -->|"max retries exceeded"| DLQ
```

## Per-Consumer Retry Infrastructure

| Service | Main Queue | Retry Queue | DLQ | Exchange | DLX |
|---|---|---|---|---|---|
| Inventory | `inventories.orders` | `inventories.orders.retry` | `inventories.orders.dlq` | `orders` | `orders.dlx` |
| Inventory | `inventories.payments` | `inventories.payments.retry` | `inventories.payments.dlq` | `payments` | `payments.dlx` |
| Order | `orders.inventories` | `orders.inventories.retry` | `orders.inventories.dlq` | `inventories` | `inventories.dlx` |
| Order | `orders.payments` | `orders.payments.retry` | `orders.payments.dlq` | `payments` | `payments.dlx` |
| Payment | `payments.inventories` | `payments.inventories.retry` | `payments.inventories.dlq` | `inventories` | `inventories.dlx` |

## Retry Behavior

- **Max retries**: 5 (tracked via `x-retry-count` header)
- **Retry TTL**: 5 seconds
- **Ack strategy**: Manual acknowledgment after successful processing
- **DLQ**: Messages that exceed max retries are routed to the dead-letter queue for
  inspection
