# High-Level Architecture

The system runs in three Docker containers backed by a shared PostgreSQL
database and Redis cache.

```mermaid
graph TB
    subgraph Clients
        Browser["Web Browser"]
        Phone["Mobile App"]
    end

    subgraph Frontend Container
        React["React SPA<br/>:3000"]
    end

    subgraph Backend Container
        WebBFF["web-bff<br/>:8080"]
        MobileBFF["mobile-bff<br/>:8081"]
        CacheWarmer["cache-warmer<br/>every 24h"]
        CartSync["cart-sync<br/>every 10s"]
        ViewSync["product-view-sync<br/>every 10s"]
    end

    subgraph Data Stores
        PG["PostgreSQL<br/>:5432"]
        Redis["Redis<br/>:6379"]
    end

    Browser --> React
    React --> WebBFF
    Phone --> MobileBFF

    WebBFF --> PG
    WebBFF --> Redis
    MobileBFF --> PG
    MobileBFF --> Redis

    CacheWarmer --> PG
    CacheWarmer --> Redis
    CartSync --> PG
    CartSync --> Redis
    ViewSync --> PG
    ViewSync --> Redis
```

## Container Breakdown

| Container | Processes | Purpose |
|---|---|---|
| Backend | 5 | Two BFF servers + three background workers |
| Frontend | 1 | React application served to browsers |
| Cache | 1 | Redis instance for caching |

The backend container demonstrates **Compute Resource Consolidation** -- all
five processes share one container because each background worker is too
lightweight to justify its own container.
