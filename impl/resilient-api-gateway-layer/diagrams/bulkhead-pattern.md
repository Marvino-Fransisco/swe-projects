# Bulkhead Pattern Diagram

```mermaid
graph TB
    subgraph Incoming["Incoming Requests"]
        R1["Request 1"]
        R2["Request 2"]
        R3["Request 3"]
        R4["Request 4"]
        R5["Request 5"]
        R6["Request 6"]
        R7["Request 7"]
        R8["Request 8"]
        R9["Request 9"]
        R10["Request 10"]
        R11["Request 11"]
    end

    subgraph Bulkhead["Bulkhead (web-backend)"]
        direction TB
        subgraph Pool["Connection Pool - max 10"]
            S1["Slot 1"]
            S2["Slot 2"]
            S3["Slot 3"]
            S4["Slot 4"]
            S5["Slot 5"]
            S6["Slot 6"]
            S7["Slot 7"]
            S8["Slot 8"]
            S9["Slot 9"]
            S10["Slot 10"]
        end
        Queue["Wait Queue (timeout: 500ms)"]
    end

    subgraph Outcomes
        Accepted["Accepted → Proceed to Circuit Breaker"]
        Rejected["Rejected → 503 CAPACITY_EXCEEDED"]
    end

    R1 --> Pool
    R2 --> Pool
    R3 --> Pool
    R4 --> Pool
    R5 --> Pool
    R6 --> Pool
    R7 --> Pool
    R8 --> Pool
    R9 --> Pool
    R10 --> Pool
    R11 --> Queue

    Pool --> Accepted
    Queue -->|timeout exceeded| Rejected

    style Pool fill:#2ecc71,stroke:#27ae60,color:#000
    style Queue fill:#f39c12,stroke:#e67e22,color:#000
    style Rejected fill:#e74c3c,stroke:#c0392b,color:#fff
    style Accepted fill:#27ae60,stroke:#1abc9c,color:#fff
```

## Bulkhead Lifecycle

```mermaid
stateDiagram-v2
    direction LR

    [*] --> Available: NewBulkhead(maxConns, queueTimeout)

    Available --> Acquiring: Acquire()
    Acquiring --> Granted: semaphore slot available
    Acquiring --> Waiting: all slots in use
    Waiting --> Granted: slot freed within timeout
    Waiting --> Rejected: queue timeout exceeded
    Granted --> InUse: request processing
    InUse --> Released: Release()
    Released --> Available: slot returned to pool
    Rejected --> [*]: return error

    note right of Available
        len(sem) < maxConns
        Semaphore-based concurrency control
    end note

    note right of Waiting
        Blocks up to queueTimeout
        (default: 500ms)
    end note
```

## Per-Service Isolation

```mermaid
graph LR
    subgraph Gateway
        BH_Mgr["BulkheadManager"]
    end

    subgraph ServiceA["web-backend"]
        BH_A["Bulkhead\nmax: 10\nqueue: 500ms"]
    end

    subgraph ServiceB["future-service"]
        BH_B["Bulkhead\nmax: 5\nqueue: 300ms"]
    end

    BH_Mgr -->|"Get('web-backend')"| BH_A
    BH_Mgr -->|"Get('future-service')"| BH_B

    style ServiceA fill:#2ecc71,stroke:#27ae60,color:#000
    style ServiceB fill:#3498db,stroke:#2980b9,color:#fff
```

> Each upstream service gets its own isolated bulkhead. A slow or overloaded
> service cannot starve resources from other services.
