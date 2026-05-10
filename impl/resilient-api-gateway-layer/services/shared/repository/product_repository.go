package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/product"
)

// productRepository implements product.ProductRepository using GORM.
type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new GORM-backed ProductRepository.
func NewProductRepository(db *gorm.DB) product.ProductRepository {
	return &productRepository{db: db}
}

// FindByID retrieves a product by its ID.
// Returns nil if not found.
func (r *productRepository) FindByID(ctx context.Context, id string) (*product.Product, error) {
	db := config.DBFromContext(ctx, r.db)
	var p product.Product
	err := db.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// FindAll retrieves a filtered and paginated list of products.
// Returns the products and the total count of matching records.
func (r *productRepository) FindAll(ctx context.Context, filter product.ProductFilter) ([]product.Product, int64, error) {
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

// SearchByName searches for products by name with pagination.
// Returns the products and the total count of matching records.
func (r *productRepository) SearchByName(ctx context.Context, query string, page, pageSize int) ([]product.Product, int64, error) {
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

// FindAllCategories retrieves all product categories.
func (r *productRepository) FindAllCategories(ctx context.Context) ([]product.Category, error) {
	db := config.DBFromContext(ctx, r.db)
	var categories []product.Category
	err := db.WithContext(ctx).Find(&categories).Error
	return categories, err
}

// AddViewCount atomically increments the view count for a product in the database.
// Uses UPDATE products SET view = view + count WHERE id = productID.
func (r *productRepository) AddViewCount(ctx context.Context, productID string, count int64) error {
	db := config.DBFromContext(ctx, r.db)
	result := db.WithContext(ctx).
		Model(&product.Product{}).
		Where("id = ?", productID).
		Update("view", gorm.Expr("view + ?", count))
	if result.Error != nil {
		return fmt.Errorf("failed to add view count for product %s: %w", productID, result.Error)
	}
	return nil
}
