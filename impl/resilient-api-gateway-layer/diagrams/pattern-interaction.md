# Pattern Interaction Diagram

This diagram shows how all resilience patterns work together in the correct
order within a single request lifecycle.

```mermaid
flowchart TB
    subgraph RequestLifecycle["Request Lifecycle"]
        direction TB
        A["1. Incoming Request\nPOST /call"]

        B["2. Bulkhead Check\nAcquire semaphore slot\nMax concurrent connections per service"]
        B -->|"Pool full + queue timeout"| B_REJECT["REJECT\n503 CAPACITY_EXCEEDED"]
        B -->|"Slot acquired"| C

        C["3. Circuit Breaker Check\nAllowRequest() based on state"]
        C -->|"OPEN (cooldown active)"| C_REJECT["REJECT\n503 SERVICE_UNAVAILABLE\nFail-fast, no upstream call"]
        C -->|"CLOSED or HALF-OPEN"| D

        D["4. Retry Loop\nUp to failureThreshold attempts\nExponential backoff + jitter"]
        D -->|"Per attempt"| E

        E["5. Timeout Enforcement\nConnection: 5s\nResponse: 10s\nDeadline: 60s total"]

        F["6. HTTP Call to Upstream\nCustom transport with timeout"]

        E --> F
        F -->|"Success (2xx/3xx)"| G["RecordSuccess()\nReset failure count\nReturn response"]
        F -->|"TCP Timeout"| H["RecordFailure()\nAbort retry immediately"]
        F -->|"Generic Timeout"| I{"Idempotent?"}
        I -->|"Yes"| J["RecordFailure()\nBackoff + Retry"]
        I -->|"No (POST/PATCH)"| H
        F -->|"5xx / 401 / 403"| K["RecordFailure()\nForce OPEN circuit\nReturn error"]
        F -->|"4xx (other)"| L["Return error\nNo CB impact"]
        F -->|"Connection error"| J

        J --> D
        H --> M["Return error to client"]
        D -->|"All retries exhausted"| N["Force OPEN circuit\nReturn MAX_RETRIES_EXCEEDED"]
    end

    subgraph Observability["Observability"]
        direction TB
        LOG["Structured Log per request\n- status_code\n- retry_count\n- circuit_state\n- bulkhead_active / bulkhead_max\n- total_latency_ms"]
        HEALTH["GET /status endpoint\n- Circuit breaker states\n- Bulkhead utilization"]
    end

    G --> LOG
    B_REJECT --> LOG
    C_REJECT --> LOG
    M --> LOG
    N --> LOG
    LOG --> HEALTH

    style B_REJECT fill:#e74c3c,stroke:#c0392b,color:#fff
    style C_REJECT fill:#e74c3c,stroke:#c0392b,color:#fff
    style G fill:#2ecc71,stroke:#27ae60,color:#000
    style K fill:#e74c3c,stroke:#c0392b,color:#fff
    style N fill:#e74c3c,stroke:#c0392b,color:#fff
    style Observability fill:#1a1a2e,stroke:#533483,color:#fff
```

## Pattern Execution Order Rationale

The order of pattern execution is critical:

1. **Bulkhead first** - Protects gateway resources. If the connection pool is
   saturated, reject immediately before spending any time on circuit breaker
   checks or retries.

2. **Circuit Breaker second** - Prevents wasted attempts. If the circuit is
   open, fail-fast without making any HTTP calls or consuming retry budget.

3. **Retry + Timeout last** - Handles transient failures. Only retry when the
   circuit is closed and there is available capacity. Retries are bounded by
   both the max attempt count and the overall request deadline.
