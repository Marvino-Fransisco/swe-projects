package cart

import "context"

// CartRepository defines the interface for cart persistence operations.
type CartRepository interface {
	// FindByUserID retrieves all cart items for a given user.
	FindByUserID(ctx context.Context, userID string) ([]Cart, error)

	// FindByUserIDAfterCursor retrieves cart items for a user using cursor-based pagination.
	// cursor is the ID of the last item from the previous page (empty string for the first page).
	// limit is the maximum number of items to return.
	// Returns the items and an error if any.
	FindByUserIDAfterCursor(ctx context.Context, userID, cursor string, limit int) ([]Cart, error)

	// FindByUserAndProduct retrieves a specific cart item by user and product.
	// Returns nil if not found.
	FindByUserAndProduct(ctx context.Context, userID, productID string) (*Cart, error)

	// Save persists a cart item (create or update).
	Save(ctx context.Context, c *Cart) error

	// Delete removes a cart item.
	Delete(ctx context.Context, c *Cart) error
}
