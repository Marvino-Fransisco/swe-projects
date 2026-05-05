# Performance Cost Optimization - Project Idea

An e-commerce application designed to learn and demonstrate three architectural
patterns: Cache-Aside, Write-Behind, and Backend For Frontends (BFF).

## Architecture Overview

Three containers:

1. **Backend Container** - Co-locates the BFF services alongside background
   workers (Compute Resource Consolidation pattern)
   - `web-bff` - Backend tailored for web clients
   - `mobile-bff` - Backend tailored for mobile clients
   - `view-counter-flusher` - Background worker: batch-flushes view counts from cache to DB
   - `cache-warmer` - Background worker: prefills product catalog into Redis on startup/schedule
   - `cart-sync-worker` - Background worker: async-persists cart data from cache to DB
2. **Frontend Container** - React user-facing application
3. **Cache Container** - Redis implementing three caching strategies

Database: PostgreSQL (shared, outside these containers)

---

## Features

### Authentication

- User registration and login
- JWT-based authentication with different token strategies per client
- This is foundational - required by most other features

### User Profile

- View and edit user profile (name, email, address, phone)
- Change password
- View order history

### Product Catalog

- Browse products by category
- Search products by name
- View product details (description, price, images, stock)
- Filter and sort products

### Product View Counter

- Track how many times a product has been viewed
- Display view count on product detail page

### Shopping Cart

- Add/remove items to cart
- Update item quantities
- View cart with total price
- Cart persists across sessions

### Checkout

- Place order from cart items
- Basic order confirmation
- View order history

---

## Pattern Mapping

### Cache-Aside Pattern

Implements two true Cache-Aside strategies (Strategies 1 and 3), plus a
Write-Behind strategy (Strategy 2) which is a related but distinct caching
pattern worth learning alongside Cache-Aside.

#### Strategy 1 - Cache-Aside: Write DB First, Then Invalidate Cache

Write to the database first, then invalidate the cache so the next read
repopulates it from fresh DB data.

**Flow:**

1. Application receives a write request
2. Write data to PostgreSQL
3. On success, invalidate the corresponding cache entry
4. On read, check cache first; on miss, read from DB and populate cache

**Applied to:**

- **User Profile** - Data integrity is critical. A user changing their
  email or address must be persisted reliably before the cache reflects it.
  Stale profile data causes real problems (wrong shipping address, wrong
  contact info). The cache serves subsequent reads to avoid hitting the DB
  on every profile view.

- **Account Settings** (password change, preferences) - Similar to profile.
  Security-sensitive data must be persisted to DB first. Cache invalidation
  ensures the next read gets fresh data.

**What to look into:**

- Cache invalidation vs cache update after DB write
- TTL (Time To Live) strategy for user data
- Handling cache update failures after successful DB write
- Race conditions when concurrent writes and reads happen

#### Strategy 2 - Write-Behind (Write-Back): Write Cache First, Then Async DB

Write to the cache first for immediate response, then asynchronously persist
to the database. This is the **Write-Behind** pattern, not Cache-Aside — the
distinction matters: Cache-Aside is a read strategy; Write-Behind is a write
strategy where the cache is the source of truth.

**Flow:**

1. Application receives a write request
2. Write data to Redis immediately (fast response to client)
3. Background worker asynchronously writes data to PostgreSQL (eventual consistency)
4. On read, always read from cache (fast); cache is authoritative during active sessions

**Applied to:**

- **Shopping Cart** - Cart updates happen frequently during a shopping
  session. Users expect instant feedback when adding/removing items. A small
  delay before the DB is updated is acceptable. If the async DB write fails,
  the cart is still in cache and can be retried. The cache is the source of
  truth during active sessions.

- **Product View Counter** - View counts update very frequently (every page
  view). Writing to DB on every view is expensive and unnecessary. Accumulate
  views in cache and batch-write to DB periodically via the
  `view-counter-flusher` worker. Exact count precision is not critical —
  eventual consistency is fine.

**What to look into:**

- The difference between Write-Behind and Cache-Aside (read vs write strategy)
- Async write-through and write-behind patterns
- Handling DB write failures when cache already has the data
- Reconciliation strategies (cache vs DB drift)
- Batch flushing view counts from cache to DB on interval
- Data durability guarantees when cache is the source of truth

#### Strategy 3 - Cache-Aside with Prefill: Read Only From Cache

Data is pre-populated into the cache by the `cache-warmer` background worker.
Reads only come from cache; the application never queries the DB at read time.
This is still Cache-Aside in spirit (the application manages the cache), but
the "miss" path is handled proactively by a worker rather than reactively by
the request handler.

**Flow:**

1. On application startup or on schedule, the `cache-warmer` loads data from
   PostgreSQL into Redis
2. On read, always read from Redis only
3. To update data, update both DB and cache (or re-trigger the warmer)

**Applied to:**

- **Product Catalog** - Products change infrequently compared to how often
  they are read. The catalog can be preloaded into cache on startup and
  refreshed periodically. Every product listing and detail page reads from
  cache only, resulting in very fast page loads.

- **App Config / Settings** - Feature flags, category lists, pricing rules.
  These are rarely changed but read on every request. Prefill on startup
  and invalidate/update only when an admin changes them.

**What to look into:**

- Cache warming / prefilling strategies
- Cold start handling (cache empty before prefill completes)
- Cache invalidation when source data changes
- Memory sizing for full catalog in cache
- Background refresh schedules vs on-demand refresh

---

### Compute Resource Consolidation

Co-locate multiple small background workers into a single container alongside
the BFFs, rather than giving each worker its own container. The point of this
pattern is avoiding idle resource waste: each worker alone is lightweight and
would leave a dedicated container largely underutilized.

**How it applies:**

The backend container runs five processes:

| Process | Role |
|---|---|
| `web-bff` | Handles web client HTTP requests |
| `mobile-bff` | Handles mobile client HTTP requests |
| `view-counter-flusher` | Periodically batch-flushes view counts from Redis to PostgreSQL |
| `cache-warmer` | Prefills and refreshes product catalog in Redis |
| `cart-sync-worker` | Async-persists cart writes from Redis to PostgreSQL |

The three background workers are the consolidation target — individually they
are too lightweight to justify dedicated containers. Co-locating them avoids
paying for three mostly-idle containers.

**What to look into:**

- Resource contention when one process consumes more CPU/memory than others
- Deployment trade-offs: updating one worker requires redeploying the whole container
- Health monitoring per process within a shared container (how do you know which one crashed?)
- Supervisor processes (e.g., `supervisord`) for managing multiple processes in one container
- When consolidation is NOT appropriate: if a worker's load spikes independently
  (e.g., catalog refresh is expensive), it may need its own container and scaling unit
- Port allocation within a single container for multiple HTTP services

---

### Backend For Frontends (BFF)

Create separate backend services tailored to specific client types.

**How it applies:**

Two BFF services, each optimized for their client. Shared business logic lives
in internal Go packages (not a separate service) to avoid code duplication
without introducing an internal network hop.

**`web-bff` (web-specific):**

- **Response shape:** Full payloads with all fields, rich data for larger screens
- **Pagination:** Offset-based pagination (page number + page size)
- **Rate limiting:** Standard limits
- **Authentication:** Standard JWT with longer token expiry

**`mobile-bff` (mobile-specific):**

- **Response shape:** Slimmed-down payloads, only essential fields to save bandwidth
- **Pagination:** Cursor-based pagination (more efficient for mobile; avoids
  skipping/duplicating items on unstable connections)
- **Rate limiting:** Stricter limits (fewer requests per minute)
- **Authentication:** Shorter token expiry with refresh token flow

**Shared logic:**

Business logic, data access, cache interactions, and validation are extracted
into internal Go packages shared by both BFFs. This avoids duplication without
creating a `main-api` service that would add network overhead and turn the BFFs
into thin proxies.

**What to look into:**

- Structuring shared Go packages vs shared internal HTTP service (trade-offs)
- Response DTO shaping per client type
- Cursor-based vs offset-based pagination implementation
- Rate limiting middleware per BFF
- Token strategy differences (expiry, refresh flow)
- API routing: how to direct web vs mobile traffic to the correct BFF

---

## Focus Areas

### Latency vs Freshness

- Cache-Aside (write-DB-first) trades write latency for stronger consistency (user profile)
- Write-Behind trades data freshness for lower write latency (shopping cart)
- Prefill/read-only-from-cache trades freshness for maximum read speed (product catalog)
- Measure and compare response times for each strategy

### Resource Utilization

- Monitor CPU/memory usage of the consolidated backend container per process
- Compare theoretical cost of 5 separate containers vs 1 consolidated container
- Identify which worker causes resource contention under load
- Track Redis memory usage across all three caching strategies

### Scaling Boundaries

- When does consolidation break down? (one process needs to scale independently)
- How does cache size grow with product catalog size?
- What happens when view counter writes exceed the batch flush worker's capacity?
- At what point does a single Redis instance become a bottleneck?
- How do you independently scale `web-bff` vs `mobile-bff` if they have different traffic profiles?

---

## Tech Stack

- **Backend:** Go
- **Frontend:** React
- **Cache:** Redis
- **Database:** PostgreSQL
- **Containerization:** Docker + Docker Compose