# Services Overview

```mermaid
graph LR
    subgraph "Backend Container (entrypoint.sh)"
        direction TB
        subgraph "BFF Services"
            WB["web-bff<br/>Gin :8080<br/>Cookie JWT<br/>Offset Pagination<br/>Full Responses"]
            MB["mobile-bff<br/>Gin :8081<br/>Bearer JWT<br/>Cursor Pagination<br/>Slim Responses"]
        end
        subgraph "Background Workers"
            CW["cache-warmer<br/>Prefills product catalog"]
            CS["cart-sync<br/>Flushes dirty carts to DB"]
            VS["product-view-sync<br/>Batch-flushes view counts"]
        end
    end

    subgraph "Shared Go Packages (shared/)"
        Domain["Domain Models<br/>(product, cart, order, user)"]
        Repo["Repositories<br/>(PostgreSQL + Redis)"]
        Svc["Domain Services<br/>(caching decorators)"]
        Config["Config<br/>(DB, Redis, Transactions)"]
    end

    WB --> Domain
    MB --> Domain
    CW --> Domain
    CS --> Domain
    VS --> Domain
    Domain --> Repo
    Domain --> Svc
    Repo --> Config
```

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
