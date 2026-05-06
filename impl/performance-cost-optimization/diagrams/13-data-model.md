# Data Model

```mermaid
erDiagram
    users {
        uuid id PK
        varchar full_name
        varchar email UK
        text address
        varchar password_hash
        timestamp created_at
        timestamp updated_at
    }

    user_preferences {
        uuid user_id PK_FK
        varchar theme
        varchar language
        timestamp created_at
        timestamp updated_at
    }

    products {
        uuid id PK
        varchar name
        text description
        decimal price
        integer stock
        bigint view
        varchar status
        timestamp created_at
        timestamp updated_at
    }

    carts {
        uuid id PK
        uuid user_id FK
        timestamp created_at
        timestamp updated_at
    }

    cart_items {
        uuid cart_id PK_FK
        uuid product_id PK_FK
        integer quantity
    }

    orders {
        uuid id PK
        uuid user_id FK
        uuid cart_id FK
        varchar status
        text failure_reason
        timestamp created_at
    }

    order_details {
        uuid order_id PK_FK
        uuid product_id PK_FK
        integer quantity
    }

    users ||--o| user_preferences : has
    users ||--|| carts : has
    users ||--o{ orders : places
    carts ||--o{ cart_items : contains
    products ||--o{ cart_items : referenced_in
    orders ||--o{ order_details : contains
    products ||--o{ order_details : referenced_in
```

## Entity Details

| Entity | Key Fields | Notes |
|---|---|---|
| User | email (unique), password_hash | Value objects: `Email`, `FullName` |
| UserPreferences | user_id (PK+FK), theme, language | One-to-one with User |
| Product | name, price, stock, view, status | Value objects: `Price`, `Stock`, `View` |
| Cart | user_id (one cart per user) | Write-Behind cached for web |
| CartItem | cart_id + product_id (composite PK) | Junction table with quantity |
| Order | user_id, cart_id, status | Status: PENDING, COMPLETED, FAILED |
| OrderDetail | order_id + product_id (composite PK) | Junction table with quantity |
