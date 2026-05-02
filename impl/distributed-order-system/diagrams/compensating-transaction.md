# Compensating Transactions - Failure Recovery

```mermaid
graph TD
    Start([Order Created]) --> Publish{Publish<br/>OrderCreated?}

    Publish -->|Success| Stock{Stock<br/>Available?}
    Publish -->|Failure| FailOrder["FailOrder<br/>(status: failed)<br/>reason: publish_fail"]

    Stock -->|Yes| Reserve["ReserveStock<br/>(status: reserved)"]
    Stock -->|No| Reject["StockRejected<br/>event"]

    Reject --> FailOrderStock["FailOrder<br/>(status: failed)<br/>reason: insufficient_inventory"]

    Reserve --> Payment{"Payment<br/>Valid?"}
    Payment -->|Yes| Success["Confirm Order<br/>Complete Reservation"]
    Payment -->|No| CancelFlow["Compensating Flow"]

    CancelFlow --> CancelOrder["Cancel Order<br/>(status: cancelled)"]
    CancelOrder --> RestoreStock["Cancel Reservation<br/>(status: cancelled)<br/>Restore Stock"]
    RestoreStock --> Done([Compensated])

    FailOrder --> DoneFail([Order Failed])
    FailOrderStock --> DoneFail2([Order Failed])
    Success --> DoneSuccess([Order Complete])

    style FailOrder fill:#f66,color:#fff
    style FailOrderStock fill:#f66,color:#fff
    style CancelOrder fill:#f96,color:#fff
    style RestoreStock fill:#f96,color:#fff
    style CancelFlow fill:#f96,color:#fff
    style Success fill:#6f6,color:#fff
    style Reserve fill:#6cf,color:#fff
```

## Compensation Triggers

| Trigger | Compensation Action | Code Location |
|---|---|---|
| Publish `OrderCreated` fails | `FailOrder` (status: failed, reason: publish_fail) | `order/internal/app/command/create_order.go` |
| Insufficient stock | `FailOrder` (status: failed, reason: insufficient_inventory) | `order/internal/app/command/fail_order.go` |
| Payment failed | `CancelReservation` + `RestoreStock` | `inventory/internal/app/command/cancel_reservation.go` |
