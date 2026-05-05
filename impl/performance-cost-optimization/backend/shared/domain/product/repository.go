package product

import "context"

// ProductFilter holds the parameters for filtering and paginating product queries.
type ProductFilter struct {
	CategoryID string
	Search     string
	Page       int
	PageSize   int
	SortBy     string
	SortOrder  string
	MinPrice   float64
	MaxPrice   float64
}

// Category represents a product category.
type Category struct {
	ID          string
	Name        string
	Description string
}

// ProductRepository defines the interface for product persistence operations.
type ProductRepository interface {
	// FindByID retrieves a product by its ID.
	// Returns nil if not found.
	FindByID(ctx context.Context, id string) (*Product, error)

	// FindAll retrieves a filtered and paginated list of products.
	// Returns the products and the total count of matching records.
	FindAll(ctx context.Context, filter ProductFilter) ([]Product, int64, error)

	// SearchByName searches for products by name with pagination.
	// Returns the products and the total count of matching records.
	SearchByName(ctx context.Context, query string, page, pageSize int) ([]Product, int64, error)

	// FindAllCategories retrieves all product categories.
	FindAllCategories(ctx context.Context) ([]Category, error)

	// GetViewCount retrieves the current view count for a product.
	GetViewCount(ctx context.Context, productID string) (int64, error)

	// IncrementViewCount increments the view count for a product by 1.
	IncrementViewCount(ctx context.Context, productID string) error
}
