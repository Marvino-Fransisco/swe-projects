# Retry and Timeout Strategy Diagram

```mermaid
flowchart TD
    Start["Incoming Request"] --> BuildURL["Build target URL from config"]
    BuildURL --> SetDeadline["Set overall request deadline\n(wall-clock limit)"]
    SetDeadline --> CreateTransport["Create custom HTTP Transport\n- Dial timeout: 5s\n- Response header timeout: 10s"]
    CreateTransport --> Loop{"Attempt\ni < maxRetries?"}

    Loop -->|No| Exhausted["All retries exhausted"]
    Exhausted --> ForceOpen["Force open circuit breaker"]
    ForceOpen --> ReturnErr["Return error to client"]

    Loop -->|Yes| CheckDeadline{"Deadline\nexceeded?"}
    CheckDeadline -->|Yes| DeadlineErr["Return TIMEOUT error"]

    CheckDeadline -->|No| CheckCB{"Circuit Breaker\nallows request?"}
    CheckCB -->|No, OPEN| FailFast["Return SERVICE_UNAVAILABLE\n(fail-fast)"]

    CheckCB -->|Yes| SendRequest["Send HTTP request\nto upstream"]

    SendRequest --> ConnError{"Connection\nerror?"}

    ConnError -->|TCP Timeout| TCPTimeout["RecordFailure()\nAbort retry (no retry for TCP timeout)"]
    TCPTimeout --> ReturnErr

    ConnError -->|Generic timeout| CheckIdempotent{"Method\nidempotent?"}
    CheckIdempotent -->|No| AbortRetry["RecordFailure()\nAbort retry"]
    AbortRetry --> ReturnErr
    CheckIdempotent -->|Yes| BackoffWait["Exponential backoff + jitter\nwait then retry"]
    BackoffWait --> RecordFail["RecordFailure()"]
    RecordFail --> Loop

    ConnError -->|Other error| OtherBackoff["Backoff + jitter\nwait then retry"]
    OtherBackoff --> RecordFail2["RecordFailure()"]
    RecordFail2 --> Loop

    SendRequest --> RespReceived["Response received"]
    RespReceived --> StatusCheck{"Status code?"}

    StatusCheck -->|2xx/3xx| Success["RecordSuccess()\nReturn response"]
    StatusCheck -->|5xx / 401 / 403| ServerErr["RecordFailure()\nForce open circuit\nReturn error immediately"]
    StatusCheck -->|Other 4xx| ClientErr["Return error immediately\nNo circuit breaker impact"]

    style FailFast fill:#e74c3c,stroke:#c0392b,color:#fff
    style DeadlineErr fill:#e67e22,stroke:#d35400,color:#fff
    style Success fill:#2ecc71,stroke:#27ae60,color:#000
    style ServerErr fill:#e74c3c,stroke:#c0392b,color:#fff
    style ClientErr fill:#f39c12,stroke:#e67e22,color:#000
```

## Exponential Backoff with Jitter

```mermaid
graph LR
    subgraph BackoffFormula
        direction TB
        B1["Attempt 0: base=1, cap=30\nbackoff = min(30, 1<<0) = 1s\nwait = 0.5s + rand(0,0.5s)"]
        B2["Attempt 1: base=1, cap=30\nbackoff = min(30, 1<<1) = 2s\nwait = 1s + rand(0,1s)"]
        B3["Attempt 2: base=1, cap=30\nbackoff = min(30, 1<<2) = 4s\nwait = 2s + rand(0,2s)"]
        B4["Attempt 3: base=1, cap=30\nbackoff = min(30, 1<<3) = 8s\nwait = 4s + rand(0,4s)"]
        B5["Attempt 4: base=1, cap=30\nbackoff = min(30, 1<<4) = 16s\nwait = 8s + rand(0,8s)"]
    end

    B1 --> B2 --> B3 --> B4 --> B5
```

## Timeout Hierarchy

```mermaid
graph TB
    subgraph Timeouts
        direction TB
        D["Request Deadline\n60s (wall-clock total)\nIncludes all retries + backoff"]
        C["Connection / Dial Timeout\n5s per attempt"]
        R["Response Header Timeout\n10s per attempt"]
        Q["Queue Timeout\n500ms (bulkhead wait)"]
    end

    D --> C
    D --> R
    Q --> D

    style D fill:#e74c3c,stroke:#c0392b,color:#fff
    style C fill:#f39c12,stroke:#e67e22,color:#000
    style R fill:#f39c12,stroke:#e67e22,color:#000
    style Q fill:#3498db,stroke:#2980b9,color:#fff
```

## Idempotency Decision Table

| HTTP Method | Idempotent | Retryable on Timeout |
| --- | --- | --- |
| GET | Yes | Yes |
| PUT | Yes | Yes |
| DELETE | Yes | Yes |
| HEAD | Yes | Yes |
| OPTIONS | Yes | Yes |
| POST | No | No (abort immediately) |
| PATCH | No | No (abort immediately) |
