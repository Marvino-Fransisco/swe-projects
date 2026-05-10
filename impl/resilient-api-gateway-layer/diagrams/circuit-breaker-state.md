# Circuit Breaker State Machine Diagram

```mermaid
stateDiagram-v2
    direction LR

    [*] --> Closed

    state Closed {
        [*] --> AcceptingRequests
        AcceptingRequests --> CountingFailures : RecordFailure
        CountingFailures --> AcceptingRequests : RecordSuccess
        CountingFailures --> ThresholdReached : failureCount >= threshold
    }

    state Open {
        [*] --> RejectingRequests
        RejectingRequests --> CooldownActive : lastFailureTime recorded
    }

    state HalfOpen {
        [*] --> ProbeRequest
        ProbeRequest --> EvaluateResult : AllowRequest returns true
    }

    Closed --> Open : EventThresholdHit\nor EventForceOpen\nor 5xx / 401 / 403
    Open --> HalfOpen : EventCooldownTimer\ncooldown elapsed
    HalfOpen --> Closed : EventSuccess\nfailureCount reset to 0
    HalfOpen --> Open : EventFailure\nsingle failure reopens circuit
```

## FSM Transition Table

```mermaid
graph TD
    subgraph States
        CLOSED["CLOSED"]
        OPEN["OPEN"]
        HALF_OPEN["HALF-OPEN"]
    end

    CLOSED -->|EventSuccess| CLOSED
    CLOSED -->|EventFailure| CLOSED
    CLOSED -->|EventThresholdHit| OPEN
    CLOSED -->|EventForceOpen| OPEN

    OPEN -->|EventCooldownTimer| HALF_OPEN
    OPEN -->|EventForceOpen| OPEN

    HALF_OPEN -->|EventSuccess| CLOSED
    HALF_OPEN -->|EventFailure| OPEN
    HALF_OPEN -->|EventForceOpen| OPEN

    style CLOSED fill:#2ecc71,stroke:#27ae60,color:#000
    style OPEN fill:#e74c3c,stroke:#c0392b,color:#fff
    style HALF_OPEN fill:#f39c12,stroke:#e67e22,color:#000
```

## State Descriptions

| State | Behavior | Color Code |
| --- | --- | --- |
| **CLOSED** | Normal operation. All requests flow through to the upstream service. Failures are counted until the threshold is reached. | Green |
| **OPEN** | Circuit is tripped. All requests are rejected immediately with a fail-fast response. Waits for the cooldown timeout to expire. | Red |
| **HALF-OPEN** | Probe state. A single request is allowed through to test if the upstream has recovered. Success closes the circuit; failure reopens it. | Yellow |
