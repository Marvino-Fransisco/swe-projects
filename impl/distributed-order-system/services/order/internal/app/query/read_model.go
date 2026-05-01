package query

import (
	"context"
	"time"
)

// OrderView is the read-side DTO returned by queries.
// It is optimized for the presentation layer and contains no domain logic.
type OrderView struct {
	ID            string
	Products      []ProductView
	Status        string
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProductView is the read-side DTO for an order product.
type ProductView struct {
	ID        string
	ProductID string
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// OrderReadModel is the port for querying order data.
// The read side (queries) uses this interface.
type OrderReadModel interface {
	GetOrderByID(ctx context.Context, id string) (OrderView, error)
	ListOrders(ctx context.Context) ([]OrderView, error)
}
