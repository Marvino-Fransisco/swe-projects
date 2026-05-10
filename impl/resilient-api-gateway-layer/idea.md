# Resilient API Gateway Layer — Build Plan

A structured checklist of everything to build for the new gateway service, organized by layer and pattern. No code — just what needs to exist and why.

---

## 1. Project Structure & Setup

- [ ] Define the service name and responsibility (e.g., `api-gateway-service`)
- [ ] Choose the tech stack for the gateway (language/framework consistent with your backend)
- [ ] Set up the project scaffold (folder structure, entry point, config files)
- [ ] Define environment config (ports, upstream service URLs, timeouts, thresholds)
- [ ] Set up dependency management (package manager / lock file)
- [ ] Add a `Dockerfile` for the gateway service
- [ ] Add the gateway to your existing `docker-compose.yml` (or equivalent orchestration)

---

## 2. Ambassador Pattern

> Acts as a proxy/helper that offloads cross-cutting concerns (logging, retries, auth) from the downstream service.

- [ ] Create a dedicated **outbound HTTP client module** that all upstream calls go through
- [ ] Move concerns out of business logic and into this client:
  - [ ] Request logging (outbound request + response)
  - [ ] Header injection (correlation IDs, tracing headers)
  - [ ] Authentication token attachment
  - [ ] Response normalization (map upstream errors to internal error types)
- [ ] Ensure the rest of the gateway calls upstream only through this client — never directly

---

## 3. Circuit Breaker Pattern

> Stops calling a failing service to allow it time to recover, instead of hammering it with requests.

- [ ] Identify which upstream service(s) the gateway calls
- [ ] Define the **three circuit states** and what triggers each:
  - [ ] `CLOSED` — normal operation, requests flow through
  - [ ] `OPEN` — failure threshold exceeded, requests are rejected immediately (fail-fast)
  - [ ] `HALF-OPEN` — after a cooldown period, allow a probe request to test recovery
- [ ] Define the **circuit breaker configuration** per upstream:
  - [ ] Failure threshold count (e.g., 5 failures → open)
  - [ ] Success threshold in half-open (e.g., 1 success → close)
  - [ ] Cooldown / reset timeout duration
- [ ] Build or integrate a circuit breaker state tracker (in-memory state machine)
- [ ] Decide the **fallback behavior** when the circuit is open:
  - [ ] Return a cached response?
  - [ ] Return a default/degraded response?
  - [ ] Return an explicit error to the caller?
- [ ] Expose circuit state in a health/status endpoint (for observability)

---

## 4. Retry Pattern

> Automatically retries failed requests before surfacing an error, useful for transient failures.

- [ ] Identify which failure types are **retryable** vs **non-retryable**:
  - [ ] Retryable: network timeout, 503 Service Unavailable, 429 Too Many Requests
  - [ ] Non-retryable: 400 Bad Request, 401 Unauthorized, 404 Not Found
- [ ] Define retry configuration per upstream call:
  - [ ] Maximum retry attempts
  - [ ] Backoff strategy: fixed delay, linear, or **exponential backoff**
  - [ ] Jitter (randomize delay to avoid thundering herd)
  - [ ] Maximum total retry duration (deadline)
- [ ] Ensure retries are **idempotent-safe** (only retry GET or idempotent operations by default)
- [ ] Track retry count in logs/metrics per request
- [ ] Decide the **retry vs fail-fast boundary** — retries must stop before the circuit breaker opens, not fight against it

---

## 5. Bulkhead Pattern

> Isolates different upstream calls into separate resource pools so one slow service can't exhaust all resources.

- [ ] Identify the distinct upstream services / call types the gateway handles
- [ ] Assign a **separate concurrency limit (thread pool / semaphore)** per upstream:
  - [ ] e.g., Service A gets max 10 concurrent connections
  - [ ] e.g., Service B gets max 5 concurrent connections
- [ ] Define **rejection behavior** when the bulkhead pool is full:
  - [ ] Fail immediately with a meaningful error
  - [ ] Queue with a short timeout, then fail
- [ ] Ensure bulkheads are scoped correctly — one slow service cannot starve the pool meant for another

---

## 6. Timeout Handling

> Every call to an upstream must have an explicit timeout — no infinite waits.

- [ ] Set **connection timeout** (time to establish a TCP connection)
- [ ] Set **read/response timeout** (time to receive the full response after connecting)
- [ ] Set **overall request deadline** (wall-clock limit for the entire operation including retries)
- [ ] Propagate timeouts via deadline headers when calling downstream services
- [ ] Handle timeout errors distinctly in logs (separate from other errors)
- [ ] Ensure timeout errors count toward the circuit breaker failure threshold

---

## 7. Routing Layer

> The gateway needs to know how to route incoming requests to the right upstream.

- [ ] Define the **route table** (which incoming path/method maps to which upstream service + endpoint)
- [ ] Apply circuit breaker + retry + bulkhead on a **per-route or per-upstream** basis
- [ ] Handle unknown routes gracefully (404 with a clear error body)

---

## 8. Error Handling & Response Normalization

- [ ] Define a **unified error response format** for all gateway errors
- [ ] Map upstream errors to gateway-level error codes:
  - [ ] Timeout → `UPSTREAM_TIMEOUT`
  - [ ] Circuit open → `SERVICE_UNAVAILABLE`
  - [ ] Retry exhausted → `MAX_RETRIES_EXCEEDED`
  - [ ] Bulkhead full → `CAPACITY_EXCEEDED`
- [ ] Never leak raw upstream error details to the caller (sanitize responses)
- [ ] Include a `correlation_id` or `request_id` in every error response

---

## 9. Observability

> Without visibility, you can't tell if the patterns are working.

- [ ] **Structured logging** on every request:
  - [ ] Upstream target, HTTP method, status code
  - [ ] Retry count, circuit state, bulkhead utilization
  - [ ] Total latency (including retries)
- [ ] **Metrics** to expose (counters + histograms):
  - [ ] Request count per upstream
  - [ ] Error rate per upstream
  - [ ] Circuit breaker state changes (opened / closed / half-open events)
  - [ ] Retry attempt distribution
  - [ ] Bulkhead rejection count
- [ ] **Health endpoint** (`GET /health`) that includes:
  - [ ] Gateway status
  - [ ] Per-upstream circuit state (`OPEN` / `CLOSED` / `HALF-OPEN`)

---

## 10. Testing Plan

- [ ] **Unit tests** for each pattern in isolation:
  - [ ] Circuit breaker state machine transitions
  - [ ] Retry logic (correct attempts, backoff timing, non-retryable passthrough)
  - [ ] Bulkhead rejection when pool is exhausted
- [ ] **Integration tests** simulating downstream failure scenarios:
  - [ ] Upstream returns 500s → verify circuit opens after threshold
  - [ ] Upstream times out → verify timeout error + retry behavior
  - [ ] Upstream recovers → verify circuit transitions to half-open then closed
- [ ] **Load / stress test** to verify bulkhead isolation:
  - [ ] Saturate one upstream → confirm other upstreams are unaffected

---

## 11. Documentation

- [ ] `README.md` for the gateway service:
  - [ ] What it does and why it exists
  - [ ] How to run it locally
  - [ ] Config reference (all env vars and their defaults)
- [ ] Architecture diagram showing: caller → gateway → upstream services, with pattern annotations
- [ ] Runbook: what to do when a circuit opens in production

---

## Pattern Interaction Map

```
Incoming Request
      │
      ▼
  [ Routing Layer ]
      │
      ▼
  [ Bulkhead ]  ← reject immediately if pool full
      │
      ▼
  [ Circuit Breaker ]  ← fail-fast if OPEN
      │
      ▼
  [ Retry + Timeout ]  ← attempt, wait, backoff, repeat
      │
      ▼
  [ Ambassador Client ]  ← actual HTTP call to upstream
      │
      ▼
  Upstream Service
```

> The order matters: bulkhead protects resources first, circuit breaker prevents wasted attempts second, retry handles transient blips third.
