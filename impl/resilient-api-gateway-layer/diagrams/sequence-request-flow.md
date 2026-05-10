# Request Processing Sequence Diagram

## Successful Request Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant R as Gin Router
    participant MW as RequestID Middleware
    participant Ctrl as CallController
    participant Svc as CallService
    participant BH as Bulkhead
    participant CB as CircuitBreaker
    participant Up as Upstream Service

    C->>R: POST /call
    R->>MW: Generate UUID
    MW->>MW: Set requestID + log in context
    MW->>Ctrl: Call(ctx)
    Ctrl->>Ctrl: Bind JSON to CallRequest
    Ctrl->>Svc: Call(req)

    Svc->>BH: Acquire()
    BH-->>Svc: OK (semaphore slot acquired)

    Svc->>CB: AllowRequest()
    CB-->>Svc: true (state is CLOSED)

    Svc->>Svc: Build target URL from config
    Svc->>Svc: Set request deadline
    Svc->>Up: HTTP Request (with timeout)

    Up-->>Svc: 200 OK + response body

    Svc->>CB: RecordSuccess()
    CB->>CB: Reset failureCount to 0
    CB->>CB: Transition: EventSuccess → CLOSED

    Svc->>BH: Release()
    Svc-->>Ctrl: CallResponse{success: true, data}
    Ctrl-->>C: 200 OK JSON
```

## Request with Retries and Circuit Breaker Opening

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant Svc as CallService
    participant BH as Bulkhead
    participant CB as CircuitBreaker
    participant Up as Upstream Service

    C->>Svc: Call(req)
    Svc->>BH: Acquire()
    BH-->>Svc: OK
    Svc->>CB: AllowRequest()
    CB-->>Svc: true (CLOSED)

    rect rgb(255, 200, 200)
        Note over Svc,Up: Attempt 1
        Svc->>Up: HTTP Request
        Up-->>Svc: Timeout Error
        Svc->>CB: RecordFailure()
        CB->>CB: failureCount++ (1/5)
        Note over Svc: Exponential backoff wait
    end

    rect rgb(255, 200, 200)
        Note over Svc,Up: Attempt 2
        Svc->>Up: HTTP Request
        Up-->>Svc: Connection Refused
        Svc->>CB: RecordFailure()
        CB->>CB: failureCount++ (2/5)
        Note over Svc: Exponential backoff wait
    end

    Note over Svc,Up: Attempts 3-5 also fail...

    rect rgb(255, 100, 100)
        Note over Svc,Up: Attempt 5 (threshold reached)
        Svc->>Up: HTTP Request
        Up-->>Svc: Timeout Error
        Svc->>CB: RecordFailure()
        CB->>CB: failureCount++ (5/5)
        CB->>CB: EventThresholdHit → OPEN
    end

    Svc->>Svc: All retries exhausted
    Svc->>CB: SetState(OPEN)
    Svc->>BH: Release()
    Svc-->>C: 503 SERVICE_UNAVAILABLE
```

## Circuit Breaker Rejecting Requests (Fail-Fast)

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant Svc as CallService
    participant BH as Bulkhead
    participant CB as CircuitBreaker

    C->>Svc: Call(req)
    Svc->>BH: Acquire()
    BH-->>Svc: OK

    Svc->>CB: AllowRequest()
    CB-->>Svc: false (state is OPEN, cooldown not elapsed)

    Svc->>BH: Release()
    Svc-->>C: 503 SERVICE_UNAVAILABLE
    Note over C,Svc: No upstream call made — fail-fast
```

## Bulkhead Rejection

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant Svc as CallService
    participant BH as Bulkhead

    C->>Svc: Call(req)
    Svc->>BH: Acquire()
    BH-->>Svc: ERROR: max concurrent connections reached
    Note over BH: Semaphore full + queue timeout exceeded

    Svc-->>C: 503 CAPACITY_EXCEEDED
    Note over C,Svc: No circuit breaker check, no upstream call
```

## Half-Open Recovery Flow

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant Svc as CallService
    participant CB as CircuitBreaker
    participant Up as Upstream Service

    Note over CB: State is OPEN, cooldown has elapsed

    C->>Svc: Call(req)
    Svc->>CB: AllowRequest()
    CB->>CB: time.Since(lastFailure) >= timeout
    CB->>CB: Transition: EventCooldownTimer → HALF-OPEN
    CB-->>Svc: true (probe allowed)

    Svc->>Up: HTTP Request (single probe)
    Up-->>Svc: 200 OK

    Svc->>CB: RecordSuccess()
    CB->>CB: Reset failureCount to 0
    CB->>CB: Transition: EventSuccess → CLOSED

    Svc-->>C: 200 OK
    Note over CB: Circuit is now CLOSED, normal operation resumes
```
