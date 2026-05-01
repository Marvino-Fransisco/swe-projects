package payment

import "context"

// Repository defines the persistence contract for Payment aggregate.
// The write side (commands) uses this interface.
type Repository interface {
	// Save persists a new payment.
	Save(ctx context.Context, payment *Payment) error

	// GetByID loads a payment by its ID.
	GetByID(ctx context.Context, id string) (*Payment, error)

	// Update persists changes to an existing payment (e.g. status change).
	Update(ctx context.Context, payment *Payment) error
}
