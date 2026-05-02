# Choreography Saga - Event Flow Diagram

```mermaid
stateDiagram-v2
    state "Order Service" as OS
    state "Inventory Service" as IS
    state "Payment Service" as PS

    state CheckoutFlow {
        [*] --> Pending : POST /api/orders
        Pending --> OrderCreated : save order
        OrderCreated --> StockCheck : orders.created

        state StockCheck {
            [*] --> ValidateStock
            ValidateStock --> StockAvailable : sufficient
            ValidateStock --> StockUnavailable : insufficient
        }

        StockAvailable --> ReservationCreated : reserve stock
        ReservationCreated --> PaymentCreated : inventories.reserved
        PaymentCreated --> AwaitPayment : save payment (pending)

        AwaitPayment --> PaymentProcessed : POST /api/payments/{id}/process
        PaymentProcessed --> OrderConfirmed : payments.succeeded
        PaymentProcessed --> ReservationConfirmed : payments.succeeded
        PaymentProcessed --> PaymentFailedEvt : payments.failed (invalid)

        StockUnavailable --> OrderFailed : inventories.rejected
        PaymentFailedEvt --> OrderCancelled : payments.failed
        PaymentFailedEvt --> StockRestored : cancel reservation

        OrderConfirmed --> [*]
        ReservationConfirmed --> [*]
        OrderFailed --> [*]
        OrderCancelled --> [*]
        StockRestored --> [*]
    }

    state "Compensations" as Comp {
        PublishFailed --> FailOrder : publish failure
        StockRejected --> FailOrder : insufficient stock
        PaymentFailedComp --> CancelOrder : payment failed
        PaymentFailedComp --> RestoreStock : restore inventory
    }
```

## Event Catalog

| Event | Routing Key | Publisher | Consumers |
|---|---|---|---|
| OrderCreated | `orders.created` | Order Service | Inventory Service |
| StockReserved | `inventories.reserved` | Inventory Service | Payment Service |
| StockRejected | `inventories.rejected` | Inventory Service | Order Service |
| PaymentSucceeded | `payments.succeeded` | Payment Service | Order Service, Inventory Service |
| PaymentFailed | `payments.failed` | Payment Service | Order Service, Inventory Service |
