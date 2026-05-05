package product

import (
	"context"
	"errors"
)

// ProductService provides domain logic for product catalog operations.
type ProductService struct {
	repo ProductRepository
}

// NewProductService creates a new ProductService with the given repository.
func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetByID retrieves a single product by its ID.
func (s *ProductService) GetByID(ctx context.Context, productID string) (*Product, error) {
	product, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	return product, nil
}

// List retrieves a filtered and paginated list of products.
// Returns the products and the total count of matching records.
func (s *ProductService) List(ctx context.Context, filter ProductFilter) ([]Product, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}

	return s.repo.FindAll(ctx, filter)
}

// Search searches for products by name with pagination.
// Returns the products and the total count of matching records.
func (s *ProductService) Search(ctx context.Context, query string, page, pageSize int) ([]Product, int64, error) {
	if query == "" {
		return nil, 0, errors.New("search query is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	return s.repo.SearchByName(ctx, query, page, pageSize)
}

// GetCategories retrieves all product categories.
func (s *ProductService) GetCategories(ctx context.Context) ([]Category, error) {
	return s.repo.FindAllCategories(ctx)
}

// GetViewCount retrieves the view count for a product.
func (s *ProductService) GetViewCount(ctx context.Context, productID string) (int64, error) {
	return s.repo.GetViewCount(ctx, productID)
}

// IncrementViewCount increments the view count for a product.
func (s *ProductService) IncrementViewCount(ctx context.Context, productID string) error {
	return s.repo.IncrementViewCount(ctx, productID)
}
