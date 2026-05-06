package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/product"
)

// ProductQueryRepository defines the interface for product query (read) operations
// that return more than one data item.
type ProductQueryRepository interface {
	// FindAll retrieves a filtered and paginated list of products.
	// Returns the products and the total count of matching records.
	FindAll(ctx context.Context, filter product.ProductFilter) ([]product.Product, int64, error)

	// SearchByName searches for products by name with pagination.
	// Returns the products and the total count of matching records.
	SearchByName(ctx context.Context, query string, page, pageSize int) ([]product.Product, int64, error)
}

type productQueryRepository struct {
	db *gorm.DB
}

// NewProductQueryRepository creates a new GORM-backed ProductQueryRepository.
func NewProductQueryRepository(db *gorm.DB) ProductQueryRepository {
	return &productQueryRepository{db: db}
}

func (r *productQueryRepository) FindAll(ctx context.Context, filter product.ProductFilter) ([]product.Product, int64, error) {
	db := config.DBFromContext(ctx, r.db)

	query := db.WithContext(ctx).Model(&product.Product{})

	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	if filter.Search != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Search+"%")
	}

	if filter.MinPrice > 0 {
		query = query.Where("price >= ?", filter.MinPrice)
	}

	if filter.MaxPrice > 0 {
		query = query.Where("price <= ?", filter.MaxPrice)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}

	sortOrder := "DESC"
	if filter.SortOrder != "" {
		sortOrder = filter.SortOrder
	}

	offset := (filter.Page - 1) * filter.PageSize
	var products []product.Product
	err := query.
		Order(fmt.Sprintf("%s %s", sortBy, sortOrder)).
		Offset(offset).
		Limit(filter.PageSize).
		Find(&products).Error

	return products, total, err
}

func (r *productQueryRepository) SearchByName(ctx context.Context, query string, page, pageSize int) ([]product.Product, int64, error) {
	db := config.DBFromContext(ctx, r.db)

	var total int64
	if err := db.WithContext(ctx).Model(&product.Product{}).
		Where("name ILIKE ?", "%"+query+"%").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	var products []product.Product
	err := db.WithContext(ctx).
		Where("name ILIKE ?", "%"+query+"%").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&products).Error

	return products, total, err
}
