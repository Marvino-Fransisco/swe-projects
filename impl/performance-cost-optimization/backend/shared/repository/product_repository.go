package repository

import (
	"context"
	"fmt"
	"strconv"

	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"shared/config"
	"shared/domain/product"
)

// productRepository implements product.ProductRepository using GORM and Redis.
type productRepository struct {
	db    *gorm.DB
	redis *goredis.Client
}

// NewProductRepository creates a new GORM-backed ProductRepository with Redis for view counts.
func NewProductRepository(db *gorm.DB, redis *goredis.Client) product.ProductRepository {
	return &productRepository{db: db, redis: redis}
}

// viewCountKey returns the Redis key used to store a product's view count.
func viewCountKey(productID string) string {
	return fmt.Sprintf("product:%s:views", productID)
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

// GetViewCount retrieves the current view count for a product from Redis.
func (r *productRepository) GetViewCount(ctx context.Context, productID string) (int64, error) {
	val, err := r.redis.Get(ctx, viewCountKey(productID)).Result()
	if err != nil {
		if err == goredis.Nil {
			return 0, nil
		}
		return 0, err
	}
	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// IncrementViewCount increments the view count for a product by 1 in Redis.
func (r *productRepository) IncrementViewCount(ctx context.Context, productID string) error {
	return r.redis.Incr(ctx, viewCountKey(productID)).Err()
}
