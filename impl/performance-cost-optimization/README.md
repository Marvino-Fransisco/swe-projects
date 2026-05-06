# Performance Cost Optimization

An e-commerce application demonstrating three cloud architecture patterns:

- **Cache-Aside Pattern** -- read-through and write-through caching strategies
- **Compute Resource Consolidation** -- co-locating lightweight workers in a
  single container to reduce idle resource waste
- **Backend For Frontends (BFF)** -- separate backend services tailored to web
  and mobile clients

The project explores the trade-offs between:

- **Latency vs Freshness** -- how caching strategies affect response times and
  data consistency
- **Resource Utilization** -- monitoring shared-container efficiency
- **Scaling Boundaries** -- when consolidation breaks down

## Tech Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.26 |
| Frontend | React |
| Database | PostgreSQL |
| Cache | Redis |
| Containerization | Docker + Docker Compose |
| Hot Reload | Air |
| ORM | GORM |
| Auth | JWT (HMAC-SHA256) |

## Quick Start

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Web BFF API | <http://localhost:8080> |
| Mobile BFF API | <http://localhost:8081> |
| Frontend | <http://localhost:3000> |
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |

---

## Diagrams

All diagrams are in the `diagrams/` directory using Mermaid syntax.

### Architecture and Overview

| Diagram | File | Description |
|---|---|---|
| High-Level Architecture | [01-high-level-architecture.md](diagrams/01-high-level-architecture.md) | Container layout and process relationships |
| Services Overview | [02-services-overview.md](diagrams/02-services-overview.md) | BFF services, workers, and shared packages |
| Code Architecture Layers | [17-code-layers.md](diagrams/17-code-layers.md) | Clean Architecture dependency flow |
| Data Model | [13-data-model.md](diagrams/13-data-model.md) | ER diagram and entity details |
| Deployment | [15-deployment.md](diagrams/15-deployment.md) | Docker Compose orchestration and startup sequence |

### Flow Diagrams

| Diagram | File | Description |
|---|---|---|
| Application Flow | [03-application-flow.md](diagrams/03-application-flow.md) | Generic request lifecycle |
| Data Flow | [04-data-flow.md](diagrams/04-data-flow.md) | Read/write paths and caching strategies |
| End-to-End Flow | [12-end-to-end-flow.md](diagrams/12-end-to-end-flow.md) | Full user journey: register to order |
| Request Lifecycle | [18-request-lifecycle.md](diagrams/18-request-lifecycle.md) | Single HTTP request through all layers |

### Per-Service Flows

| Diagram | File | Description |
|---|---|---|
| Authentication | [06-auth-flow.md](diagrams/06-auth-flow.md) | Register, login, and token validation |
| Auth Comparison | [16-auth-comparison.md](diagrams/16-auth-comparison.md) | Web cookies vs mobile bearer tokens |
| Product Catalog | [07-product-flow.md](diagrams/07-product-flow.md) | List, search, and detail |
| Shopping Cart | [08-cart-flow.md](diagrams/08-cart-flow.md) | Web (cached) vs mobile (direct DB) |
| Checkout | [09-checkout-flow.md](diagrams/09-checkout-flow.md) | Place order with stock validation |
| Profile | [10-profile-flow.md](diagrams/10-profile-flow.md) | View, update, and change password |

### Caching and Workers

| Diagram | File | Description |
|---|---|---|
| Caching Strategies | [05-caching-strategies.md](diagrams/05-caching-strategies.md) | All 4 strategies with sequence diagrams |
| Worker Detail Flows | [11-worker-flows.md](diagrams/11-worker-flows.md) | Flowcharts for each background worker |
| Patterns Summary | [14-patterns-summary.md](diagrams/14-patterns-summary.md) | Pattern overview and trade-offs |

---

## API Reference

Both BFF services expose identical URL paths under `/api/v1`, differing in auth
method, response shape, and pagination strategy.

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | Public | Register new user |
| POST | `/api/v1/auth/login` | Public | Login and receive tokens |
| POST | `/api/v1/auth/refresh` | Public | Refresh access token |
| POST | `/api/v1/auth/logout` | Public | Clear tokens |

### Products

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/products` | Optional | List products with filters |
| GET | `/api/v1/products/search?q=` | Optional | Search by name |
| GET | `/api/v1/products/categories` | Optional | List categories |
| GET | `/api/v1/products/:id` | Optional | Product detail |
| POST | `/api/v1/products/:id/view` | Optional | Track product view |

### Cart

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/cart` | Required | Get cart items |
| POST | `/api/v1/cart/items` | Required | Add item to cart |
| PUT | `/api/v1/cart/items/:productId` | Required | Update quantity |
| DELETE | `/api/v1/cart/items/:productId` | Required | Remove item |

### Checkout

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/checkout/orders` | Required | Place order |
| GET | `/api/v1/checkout/orders` | Required | Order history |
| GET | `/api/v1/checkout/orders/:id` | Required | Order detail |

### Profile

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/api/v1/profile` | Required | Get profile |
| PUT | `/api/v1/profile` | Required | Update profile |
| PUT | `/api/v1/profile/password` | Required | Change password |

---

## BFF Comparison

| Aspect | web-bff | mobile-bff |
|---|---|---|
| Port | 8080 | 8081 |
| Auth | HTTP-only cookies | Bearer token header |
| Product list | Full `Product` objects | `ProductSummary` (id, name, price) |
| Cart pagination | Offset (`page`, `page_size`) | Cursor (`cursor`, `page_size`) |
| Cart response | Raw `CartItem` | Enriched with `product_name`, `price` |
| Order response | Full `Order` | Slim `OrderResponse` |
| Cart caching | Write-Behind (Redis + async DB) | Direct DB |
| Profile caching | Cache-Aside (DB write + cache invalidate) | Direct DB |

---

## Redis Key Reference

| Key | Type | TTL | Used By | Strategy |
|---|---|---|---|---|
| `products` | JSON string | 7 days | cache-warmer, product reads | Prefill |
| `user:<uuid>` | Hash | 7 days | profile reads/writes | Cache-Aside |
| `cart:<userId>` | JSON string | 24 hours | web-bff cart operations | Write-Behind |
| `cart:dirty` | Set | -- | cart-sync worker | Write-Behind flag |
| `product:view_counts` | Hash | -- | view tracking, view-sync | Batch Counter |

---

## Project Structure

```
performance-cost-optimization/
├── docker-compose.yml
├── README.md
├── idea.md
├── diagrams/                       # All architecture and flow diagrams
├── frontend/
│   └── Dockerfile
├── backend/
│   ├── Dockerfile
│   ├── entrypoint.sh              # Launches all 5 processes
│   ├── shared/                    # Shared Go packages
│   │   ├── cmd/                   # DB migration + seeding
│   │   ├── config/                # PostgreSQL, Redis, transaction configs
│   │   ├── domain/                # Domain models + services
│   │   │   ├── product/           # Product entity, value objects, cache repo
│   │   │   ├── cart/              # Cart entity, cached service decorator
│   │   │   ├── order/             # Order entity, order details
│   │   │   └── user/              # User entity, preferences, cache repo
│   │   ├── middleware/            # Shared auth middleware
│   │   ├── repository/           # PostgreSQL + Redis implementations
│   │   ├── util/                  # JWT, bcrypt helpers
│   │   └── workers/              # Background workers
│   │       ├── cache-warmer.go
│   │       ├── cart-sync.go
│   │       └── product-view-sync.go
│   └── services/
│       ├── web-backend/           # Web BFF service
│       │   ├── bootstrap/         # DI wiring
│       │   ├── controller/        # HTTP handlers
│       │   ├── middleware/        # Cookie auth
│       │   ├── repository/       # Read-optimized queries
│       │   └── usecases/         # Business logic per feature
│       ├── mobile-backend/        # Mobile BFF service
│       │   ├── bootstrap/
│       │   ├── controller/
│       │   ├── middleware/        # Bearer token auth
│       │   ├── repository/
│       │   └── usecases/
│       └── workers/               # Worker process entrypoint
│           └── main.go
```
