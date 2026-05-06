package product

import (
	"context"
	"errors"
	"fmt"
)

// ProductService provides domain logic for product catalog operations.
type ProductService struct {
	repo      ProductRepository
	cacheRepo ProductCacheRepository
}

// NewProductService creates a new ProductService with the given repositories.
func NewProductService(repo ProductRepository, cacheRepo ProductCacheRepository) *ProductService {
	return &ProductService{repo: repo, cacheRepo: cacheRepo}
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

// GetCategories retrieves all product categories.
func (s *ProductService) GetCategories(ctx context.Context) ([]Category, error) {
	return s.repo.FindAllCategories(ctx)
}

// IncrementViewCount increments the view count for a product in the cache.
func (s *ProductService) IncrementViewCount(ctx context.Context, productID string) error {
	return s.cacheRepo.IncrementViewCount(ctx, productID)
}

// GetViewCount returns the total view count for a product (DB view + cache counter).
func (s *ProductService) GetViewCount(ctx context.Context, productID string) (int64, error) {
	product, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return 0, fmt.Errorf("failed to get product from db: %w", err)
	}
	if product == nil {
		return 0, errors.New("product not found")
	}

	dbView := product.View.Int64()

	cachedView, err := s.cacheRepo.GetViewCount(ctx, productID)
	if err != nil {
		return 0, fmt.Errorf("failed to get cached view count: %w", err)
	}

	return dbView + cachedView, nil
}

// FlushAllViewCounts reads all cached view counters, atomically updates each product's
// view count in the database (SET view = view + counter), then resets all counters in cache.
func (s *ProductService) FlushAllViewCounts(ctx context.Context) error {
	viewCounts, err := s.cacheRepo.GetAllViewCounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all view counts from cache: %w", err)
	}

	for productID, count := range viewCounts {
		if count <= 0 {
			continue
		}
		if err := s.repo.AddViewCount(ctx, productID, count); err != nil {
			return fmt.Errorf("failed to add view count for product %s: %w", productID, err)
		}
	}

	if err := s.cacheRepo.ResetAllViewCounts(ctx); err != nil {
		return fmt.Errorf("failed to reset view counts in cache: %w", err)
	}

	return nil
}
