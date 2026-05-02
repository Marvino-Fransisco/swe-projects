# Competing Consumers - Inventory Workers

```mermaid
graph TB
    subgraph RabbitMQ
        Q["inventories.orders<br/>queue"]
    end

    subgraph InventoryService ["Inventory Service"]
        W1["Worker 1<br/>(goroutine)"]
        W2["Worker 2<br/>(goroutine)"]
        W3["Worker 3<br/>(goroutine)"]
    end

    subgraph PostgreSQL
        Inv["inventories table"]
        Res["inventory_reservations table"]
    end

    Q -->|"message 1"| W1
    Q -->|"message 2"| W2
    Q -->|"message 3"| W3

    W1 -->|"SELECT FOR UPDATE<br/>(row lock)"| Inv
    W2 -->|"SELECT FOR UPDATE<br/>(row lock)"| Inv
    W3 -->|"SELECT FOR UPDATE<br/>(row lock)"| Inv

    W1 -->|"INSERT"| Res
    W2 -->|"INSERT"| Res
    W3 -->|"INSERT"| Res
```

## How It Works

1. Three goroutines consume from the same `inventories.orders` queue
2. RabbitMQ delivers each message to exactly one worker (round-robin by default)
3. Each worker opens a database transaction and uses `SELECT FOR UPDATE` to acquire
   row-level locks on the relevant inventory rows
4. If two workers try to reserve the same product, the second one blocks until the
   first commits or rolls back
5. Manual acknowledgment ensures at-least-once delivery semantics
