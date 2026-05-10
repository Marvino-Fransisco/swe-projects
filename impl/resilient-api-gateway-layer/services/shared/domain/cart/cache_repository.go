package cart

import "context"

// CartCacheRepository defines the interface for cart cache operations.
type CartCacheRepository interface {
	// GetByUserID retrieves the user's cart from cache.
	// Returns nil if not found in cache.
	GetByUserID(ctx context.Context, userID string) (*Cart, error)

	// GetCartDirtyMembers retrieves the list of userIDs whose carts are marked dirty in cache.
	GetCartDirtyMembers(ctx context.Context) ([]string, error)

	// Set persists the user's cart to cache with a TTL.
	Set(ctx context.Context, userID string, c *Cart) error

	// SetDirty marks a user's cart as dirty in cache with a TTL.
	SetDirty(ctx context.Context, userID string) error

	// Delete removes the user's cart from cache.
	Delete(ctx context.Context, userID string) error

	DeleteCartDirtyMember(ctx context.Context, userID string) error
}
