# Code Architecture Layers

Shows the Clean Architecture / DDD layering within the codebase and how
dependencies flow inward.

```mermaid
graph TB
    subgraph "Service Layer (services/)"
        Main["main.go"]
        Router["Router (Gin)"]
        Controller["Controller"]
        UseCase["Use Case"]
        SvcRepo["Query Repository"]
        SvcMW["Middleware"]
    end

    subgraph "Shared Layer (shared/)"
        Domain["Domain Entities<br/>& Value Objects"]
        DomainSvc["Domain Services<br/>(caching decorators)"]
        DomainRepo["Repository Interfaces"]
        InfraRepo["Repository Implementations<br/>(PostgreSQL + Redis)"]
        Config["Config<br/>(DB, Redis, Transactions)"]
        Workers["Workers"]
        Util["Util (JWT, Hash)"]
        SharedMW["Middleware (auth)"]
    end

    subgraph "Infrastructure"
        PG["PostgreSQL"]
        Redis["Redis"]
    end

    Main --> Router
    Router --> SvcMW
    Router --> Controller
    Controller --> UseCase
    UseCase --> DomainSvc
    UseCase --> SvcRepo
    DomainSvc --> DomainRepo
    DomainRepo -.->|"implemented by"| InfraRepo
    InfraRepo --> Config
    Domain --> DomainSvc
    Workers --> DomainSvc
    Workers --> InfraRepo
    SvcMW --> SharedMW
    SharedMW --> Util
    Config --> PG
    Config --> Redis

    style Domain fill:#e1f5fe
    style DomainSvc fill:#e8f5e9
    style InfraRepo fill:#fff3e0
```

## Dependency Rule

```mermaid
graph LR
    Controller -->|"depends on"| UseCase
    UseCase -->|"depends on"| DomainSvc["Domain Service"]
    UseCase -->|"depends on"| DomainRepo["Repository Interface"]
    DomainSvc -->|"depends on"| DomainRepo
    DomainRepo -.->|"implemented by"| InfraRepo["Infrastructure Repo"]
    InfraRepo -->|"depends on"| Config

    style UseCase fill:#e8f5e9
    style DomainRepo fill:#e1f5fe
    style InfraRepo fill:#fff3e0
```

Dependencies point inward. The domain layer has zero external dependencies.
Controllers and infrastructure depend on domain abstractions, never the
other way around.
