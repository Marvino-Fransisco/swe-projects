package product

import "context"

// ProductCacheRepository defines the interface for product view count cache operations.
type ProductCacheRepository interface {
	// IncrementViewCount increments the view count for a product by 1 in cache.
	IncrementViewCount(ctx context.Context, productID string) error

	// GetViewCount retrieves the cached view count for a product.
	// Returns 0 if not found in cache.
	GetViewCount(ctx context.Context, productID string) (int64, error)

	// GetAllViewCounts retrieves all product view counts from cache.
	// Returns a map of productID -> view count.
	GetAllViewCounts(ctx context.Context) (map[string]int64, error)

	// ResetAllViewCounts resets or deletes all product view counters in cache.
	ResetAllViewCounts(ctx context.Context) error
}
