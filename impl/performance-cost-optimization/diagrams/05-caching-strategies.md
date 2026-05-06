# Caching Strategies Detail

## Strategy 1 -- Cache-Aside: Write DB First, Then Invalidate Cache

Applied to **User Profile** updates. Data integrity is critical -- changes must
be persisted to the database before the cache is updated.

```mermaid
sequenceDiagram
    actor Client
    participant BFF
    participant DB as PostgreSQL
    participant Cache as Redis

    Note over Client, Cache: Write Flow

    Client->>BFF: PUT /api/v1/profile
    BFF->>DB: BEGIN TRANSACTION
    BFF->>DB: UPDATE users SET ...
    DB-->>BFF: OK
    BFF->>DB: COMMIT
    BFF->>Cache: PIPELINE (DELETE + HSET)
    Cache-->>BFF: OK
    BFF-->>Client: 200 Updated profile

    Note over Client, Cache: Read Flow

    Client->>BFF: GET /api/v1/profile
    BFF->>Cache: GET user:uuid
    alt Cache Hit
        Cache-->>BFF: Cached user data
    else Cache Miss
        Cache-->>BFF: nil
        BFF->>DB: SELECT * FROM users
        DB-->>BFF: User data
        BFF->>Cache: HSET user:uuid (TTL 7d)
    end
    BFF-->>Client: 200 Profile data
```

**Redis keys:** `user:<uuid>` (hash), TTL 7 days

---

## Strategy 2 -- Write-Behind: Write Cache First, Async DB

Applied to **Shopping Cart**. The cache is the source of truth during active
sessions, and the `cart-sync` worker persists to the database asynchronously.

```mermaid
sequenceDiagram
    actor Client
    participant BFF
    participant Cache as Redis
    participant Worker as Background Worker
    participant DB as PostgreSQL

    Note over Client, DB: Shopping Cart Write

    Client->>BFF: POST /api/v1/cart/items
    BFF->>Cache: SET cart:userId (TTL 24h)
    BFF->>Cache: SADD cart:dirty userId
    Cache-->>BFF: OK
    BFF-->>Client: 201 Item added (fast response)

    Note over Client, DB: Background Sync (every 10s)

    Worker->>Cache: SMEMBERS cart:dirty
    Cache-->>Worker: [userId1, userId2, ...]
    loop For each dirty user
        Worker->>Cache: GET cart:userId
        Cache-->>Worker: Cart data
        Worker->>DB: BEGIN TRANSACTION
        Worker->>DB: UPSERT cart_items
        DB-->>Worker: OK
        Worker->>DB: COMMIT
        Worker->>Cache: SREM cart:dirty userId
    end
```

**Redis keys:**

- `cart:<userId>` (JSON), TTL 24h
- `cart:dirty` (set of user IDs with pending changes)

---

## Strategy 2b -- Product View Counter (Batch Flush)

```mermaid
sequenceDiagram
    actor Client
    participant BFF
    participant Cache as Redis
    participant Worker as product-view-sync
    participant DB as PostgreSQL

    Client->>BFF: POST /api/v1/products/:id/view
    BFF->>Cache: HINCRBY product:view_counts productId 1
    Cache-->>BFF: OK
    BFF-->>Client: 200 Tracked

    Note over Cache, DB: Flush every 10 seconds

    Worker->>Cache: HGETALL product:view_counts
    Cache-->>Worker: {prod1: 47, prod2: 12, ...}
    loop For each product
        Worker->>DB: UPDATE products SET view = view + count
    end
    Worker->>Cache: DEL product:view_counts
```

**Redis keys:** `product:view_counts` (hash of productId to count)

---

## Strategy 3 -- Cache-Aside with Prefill: Read Only from Cache

Applied to **Product Catalog**. The `cache-warmer` worker preloads all products
into Redis on startup and every 24 hours.

```mermaid
sequenceDiagram
    participant Worker as cache-warmer
    participant DB as PostgreSQL
    participant Cache as Redis

    Note over Worker, Cache: On startup + every 24h

    Worker->>DB: SELECT * FROM products
    DB-->>Worker: All products
    Worker->>Cache: SET products (JSON, TTL 7d)
    Cache-->>Worker: OK
```

**Redis keys:** `products` (JSON-encoded slice), TTL 7 days
