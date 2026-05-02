# Claim Check Pattern - Large Payload Optimization

```mermaid
sequenceDiagram
    participant Order as Order Service
    participant Redis as Redis
    participant RMQ as RabbitMQ
    participant Inv as Inventory Service

    Order->>Redis: Store payload<br/>key: claim-check:orders:{id}<br/>TTL: 1 hour
    Redis-->>Order: OK

    Order->>RMQ: Publish lightweight message<br/>{orderId, claimCheckKey}<br/>routing key: orders.created
    RMQ-->>Inv: Deliver message

    Inv->>Redis: Fetch payload<br/>key: claim-check:orders:{id}
    Redis-->>Inv: Full order payload

    Inv->>Redis: Delete claim check key
    Redis-->>Inv: OK

    Inv->>Inv: Process order<br/>(ReserveStock)
```

## Why Claim Check

Order payloads contain multiple products with quantities. Rather than serializing
the full payload through RabbitMQ, the Order Service stores it in Redis and only
passes a lightweight reference message. The Inventory Service fetches the full
payload from Redis when it processes the message, then removes the key.

- **Storage**: Redis key `claim-check:orders:{orderId}`
- **TTL**: 1 hour (automatic cleanup)
- **Removal**: Consumer deletes key after fetching
