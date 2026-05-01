package order

import "context"

// Repository defines the persistence contract for Order aggregate.
// The write side (commands) uses this interface.
type Repository interface {
	// Save persists a new order.
	Save(ctx context.Context, order *Order) error

	// GetByID loads an order by its ID, including its products.
	GetByID(ctx context.Context, id string) (*Order, error)

	// Update persists changes to an existing order (e.g. status change).
	Update(ctx context.Context, order *Order) error
}
