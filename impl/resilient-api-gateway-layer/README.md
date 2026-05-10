# Resilient API Gateway Layer

A Go-based API gateway that implements enterprise-grade resilience patterns to
protect downstream services from cascading failures. The gateway acts as an
**Ambassador** proxy that offloads cross-cutting concerns — circuit breaking,
bulkhead isolation, retry with backoff, and timeout management — from both the
client and the upstream services.

## Architecture Overview

See [diagrams/architecture.md](diagrams/architecture.md) for the full
high-level architecture diagram.

```markdown
Client
  │
  ▼
┌──────────────────────────────────────────────────┐
│              Ambassador Gateway (:6969)          │
│                                                  │
│  Request ID → Router → Controller → Service      │
│                                     │            │
│                          ┌──────────┼──────────┐ │
│                          │ Resilience Layer    │ │
│                          │                     │ │
│                          │  1. Bulkhead        │ │
│                          │  2. Circuit Breaker │ │
│                          │  3. Retry + Timeout │ │
│                          └──────────┼──────────┘ │
│                                     │            │
│                            HTTP Client           │
└─────────────────────────────────┬────────────────┘
                                  │
                                  ▼
                        ┌───────────────────┐
                        │  Upstream Service │
                        │  web-backend      │
                        │  (:8080)          │
                        └────────┬──────────┘
                                 │
                        ┌────────┴──────────┐
                        │  PostgreSQL       │
                        │  Redis            │
                        └───────────────────┘
```

## Resilience Patterns

This project implements four interconnected resilience patterns. The execution
order matters — each layer protects against a different failure mode.

See [diagrams/pattern-interaction.md](diagrams/pattern-interaction.md) for the
full pattern interaction diagram.

### 1. Ambassador Pattern

The gateway acts as a client-side ambassador proxy. All outbound calls to
upstream services pass through a single controlled path that handles:

- Request logging with structured fields
- Correlation ID (`X-Request-ID`) generation and propagation
- Service name header injection (`X-Service-Name`)
- Deadline propagation (`X-Request-Timeout`)
- Response normalization and error sanitization

### 2. Bulkhead Pattern

Isolates concurrent connections per upstream service using a semaphore-based
pool. Prevents one slow or overloaded service from exhausting gateway
resources.

See [diagrams/bulkhead-pattern.md](diagrams/bulkhead-pattern.md) for details.

- **Per-service isolation** — each upstream gets its own connection pool
- **Queue with timeout** — requests wait up to `queue_timeout` (default 500ms)
  for a slot before being rejected
- **Fail-fast rejection** — returns `CAPACITY_EXCEEDED` immediately when pool
  is saturated

### 3. Circuit Breaker Pattern

Stops calling a failing upstream service to give it time to recover, instead
of continuously hammering it with requests.

See [diagrams/circuit-breaker-state.md](diagrams/circuit-breaker-state.md) for
the full state machine diagram.

- **CLOSED** — normal operation, requests flow through, failures are counted
- **OPEN** — failure threshold exceeded, all requests rejected immediately
  (fail-fast)
- **HALF-OPEN** — after cooldown expires, allows a single probe request to
  test recovery

The circuit breaker is implemented as a finite state machine (`FSM`) with
explicit transition rules. See `pkg/fsm.go` for the FSM engine and
`pkg/circuit-breaker.go` for the circuit breaker logic.

### 4. Retry Pattern

Automatically retries failed requests with exponential backoff and jitter for
transient failures. Key behaviors:

See [diagrams/retry-timeout-flow.md](diagrams/retry-timeout-flow.md) for the
full retry and timeout strategy diagram.

- **Idempotency-aware** — only retries `GET`, `PUT`, `DELETE`, `HEAD`,
  `OPTIONS`; immediately aborts on timeout for `POST` and `PATCH`
- **TCP timeout abort** — TCP connection timeouts are not retried (indicates
  network-level issue, not transient)
- **Exponential backoff with jitter** — `backoff = min(30, 1 << attempt)`
  with random jitter to prevent thundering herd
- **Request deadline** — overall wall-clock limit (default 60s) that caps
  total time including all retries and backoff waits
- **5xx immediate trip** — upstream 5xx, 401, or 403 responses immediately
  force the circuit breaker open without retry

## Project Structure

```markdown
resilient-api-gateway-layer/
├── README.md
├── docker-compose.yml
├── diagrams/                        # Architecture and design diagrams
│   ├── architecture.md
│   ├── bulkhead-pattern.md
│   ├── circuit-breaker-state.md
│   ├── component-diagram.md
│   ├── error-handling.md
│   ├── pattern-interaction.md
│   ├── retry-timeout-flow.md
│   └── sequence-request-flow.md
└── services/
    ├── ambassador/                   # Gateway service
    │   ├── main.go                   # Entry point, wiring
    │   ├── config.yaml               # Per-service resilience config
    │   ├── .env                      # Environment variables
    │   ├── configs/
    │   │   └── logging.go            # Logger setup (dev/prod)
    │   ├── controllers/
    │   │   ├── call.go               # POST /call handler
    │   │   └── health.go             # GET /status handler
    │   ├── dtos/
    │   │   ├── call.go               # Request/response DTOs
    │   │   ├── error.go              # Error codes and ErrorResponse
    │   │   ├── health.go             # Health status DTOs
    │   │   └── header.go             # Header propagation DTOs
    │   ├── lib/
    │   │   ├── env.go                # .env file loader
    │   │   └── yaml.go               # config.yaml loader
    │   ├── middlewares/
    │   │   └── create_request_id.go  # UUID + log context middleware
    │   ├── pkg/
    │   │   ├── bulkhead.go           # Semaphore-based bulkhead
    │   │   ├── bulkhead-manager.go   # Per-service bulkhead registry
    │   │   ├── circuit-breaker.go    # Circuit breaker with FSM
    │   │   ├── circuit-breaker-manager.go  # Per-service CB registry
    │   │   └── fsm.go               # Generic finite state machine
    │   ├── routers/
    │   │   └── router.go             # Gin route definitions
    │   └── services/
    │       ├── call.go               # Core request proxy logic
    │       └── health.go             # Health aggregation
    └── web-backend/                  # Downstream e-commerce service
        └── main.go                   # Runs on :8080
```

## API Reference

### POST /call

Proxies a request to a configured upstream service through the resilience
layer.

**Request Body:**

```json
{
  "target_service_name": "web-backend",
  "method": "GET",
  "url": "/api/v1/products",
  "headers": {},
  "body": null
}
```

**Success Response (200):**

```json
{
  "success": true,
  "data": { ... }
}
```

**Error Response (any):**

```json
{
  "success": false,
  "error_code": "SERVICE_UNAVAILABLE",
  "message": "circuit breaker is open for service: web-backend",
  "request_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### GET /status

Returns the current state of all circuit breakers and bulkheads.

**Response (200):**

```json
{
  "success": true,
  "circuit_breakers": [
    {
      "service_name": "web-backend",
      "state": "closed",
      "failure_count": 0,
      "failure_threshold": 5
    }
  ],
  "bulkheads": [
    {
      "service_name": "web-backend",
      "active_connections": 3,
      "max_connections": 10
    }
  ]
}
```

### Error Codes

| Error Code | HTTP Status | Description |
| --- | --- | --- |
| `BAD_REQUEST` | 400 | Invalid request body or parameters |
| `UNAUTHORIZED` | 401 | Authentication required |
| `FORBIDDEN` | 403 | Access denied |
| `RESOURCE_NOT_FOUND` | 404 | Route or upstream resource not found |
| `TIMEOUT` | 504 | Request or deadline timeout |
| `TCP_CONNECTION_TIMEOUT` | 504 | Failed to establish TCP connection |
| `SERVICE_UNAVAILABLE` | 502 | Upstream error or circuit breaker open |
| `INTERNAL_SERVER_ERROR` | 500 | Unexpected gateway error |
| `MAX_RETRIES_EXCEEDED` | 502 | All retry attempts exhausted |
| `CAPACITY_EXCEEDED` | 502 | Bulkhead connection pool full |

See [diagrams/error-handling.md](diagrams/error-handling.md) for the full
error handling flow.

## Configuration

### Environment Variables (.env)

| Variable | Description | Default |
| --- | --- | --- |
| `PORT` | Gateway listen port | `6969` |
| `ENVIRONMENT` | `dev` for text logs, `prod` for JSON logs | `dev` |

### Service Configuration (config.yaml)

```yaml
services:
  web-backend:
    host: web-backend_devcontainer-app-1
    port: 8080
    timeout: 5            # Circuit breaker cooldown (seconds)
    threshold: 5          # Failure threshold to trip circuit + max retries
    max_connections: 10   # Bulkhead concurrent connection limit
    queue_timeout: 500    # Bulkhead queue wait timeout (milliseconds)
    request_deadline: 60  # Overall request deadline (seconds)
```

| Field | Description |
| --- | --- |
| `host` | Upstream service hostname |
| `port` | Upstream service port |
| `timeout` | Circuit breaker cooldown duration before transitioning to HALF-OPEN |
| `threshold` | Number of failures before circuit opens; also used as max retry count |
| `max_connections` | Maximum concurrent connections in the bulkhead pool |
| `queue_timeout` | Maximum time to wait for a bulkhead slot (ms) |
| `request_deadline` | Wall-clock timeout for the entire request including retries (s) |

## Timeout Hierarchy

Three layers of timeout ensure no request hangs indefinitely:

| Timeout | Duration | Scope |
| --- | --- | --- |
| **Queue Timeout** | 500ms | Wait for bulkhead slot |
| **Connection Timeout** | 5s | TCP dial per attempt |
| **Response Timeout** | 10s | Response header per attempt |
| **Request Deadline** | 60s | Total wall-clock including all retries |

## Diagrams

All diagrams are written in Mermaid syntax and located in the `diagrams/`
directory:

| Diagram | File | Description |
| --- | --- | --- |
| High-Level Architecture | [architecture.md](diagrams/architecture.md) | System overview and request flow |
| Circuit Breaker State Machine | [circuit-breaker-state.md](diagrams/circuit-breaker-state.md) | FSM states and transitions |
| Request Sequence Diagrams | [sequence-request-flow.md](diagrams/sequence-request-flow.md) | Success, retry, fail-fast, recovery flows |
| Component / Class Diagram | [component-diagram.md](diagrams/component-diagram.md) | Go types and their relationships |
| Bulkhead Pattern | [bulkhead-pattern.md](diagrams/bulkhead-pattern.md) | Connection pool isolation |
| Retry and Timeout Strategy | [retry-timeout-flow.md](diagrams/retry-timeout-flow.md) | Backoff, idempotency, timeout hierarchy |
| Error Handling | [error-handling.md](diagrams/error-handling.md) | Error normalization and sanitization |
| Pattern Interaction | [pattern-interaction.md](diagrams/pattern-interaction.md) | How all patterns work together |

## Getting Started

### Prerequisites

- Go 1.22+
- Docker and Docker Compose

### Run with Docker Compose

```bash
docker compose up -d postgres
```

### Run the Gateway

```bash
cd services/ambassador
cp .env.example .env
go run main.go
```

### Run the Backend

```bash
cd services/web-backend
go run main.go
```

### Example Request

```bash
curl -X POST http://localhost:6969/call \
  -H "Content-Type: application/json" \
  -d '{
    "target_service_name": "web-backend",
    "method": "GET",
    "url": "/api/v1/products"
  }'
```

### Check Gateway Status

```bash
curl http://localhost:6969/status
```

## Key Design Decisions

- **Bulkhead before Circuit Breaker** — resource protection takes priority over
  failure detection. A saturated bulkhead rejects before checking circuit state.
- **Circuit Breaker before Retry** — fail-fast prevents wasted retry attempts
  when the circuit is already open.
- **Retry is idempotency-safe** — non-idempotent methods (`POST`, `PATCH`) are
  never retried on timeout to prevent duplicate side effects.
- **5xx immediately trips the circuit** — upstream server errors, auth failures
  (401, 403) force the circuit open without waiting for the failure threshold,
  since these indicate a known-downstream problem.
- **4xx does not affect circuit state** — client errors (400, 404) are passed
  through without counting as failures, since they are not indicators of
  upstream health.
- **TCP timeouts are not retried** — a TCP connection timeout indicates a
  network-level issue, not a transient application failure.
- **Error responses are sanitized** — upstream error details are never leaked
  to the client. All responses use the gateway error format with a
  `request_id` for tracing.

## Tech Stack

- **Language:** Go
- **HTTP Framework:** Gin
- **Logging:** Logrus (structured logging, dev/prod formatters)
- **Resilience:** Custom implementations (no external resilience library)
- **Database:** PostgreSQL (via Docker Compose)
- **Containerization:** Docker, Docker Compose
