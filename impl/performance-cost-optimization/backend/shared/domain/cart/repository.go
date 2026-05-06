package cart

import "context"

// CartRepository defines the interface for cart persistence operations.
type CartRepository interface {
	// FindByUserID retrieves the cart for a given user, preloading products.
	// Returns nil if not found.
	FindByUserID(ctx context.Context, userID string) (*Cart, error)

	// Save persists a cart (create or update), including its product associations.
	Save(ctx context.Context, c *Cart) error

	// Delete removes a cart.
	Delete(ctx context.Context, c *Cart) error

	ReplaceCart(ctx context.Context, c *Cart) error
}
