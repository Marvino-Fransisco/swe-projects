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

	// FindAllCategories retrieves all product categories.
	FindAllCategories(ctx context.Context) ([]Category, error)

	// AddViewCount atomically increments the view count for a product in the database
	// using UPDATE products SET view = view + count WHERE id = productID.
	AddViewCount(ctx context.Context, productID string, count int64) error
}
