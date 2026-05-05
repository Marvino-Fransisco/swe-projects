package repository

import (
	"context"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/cart"
)

// cartRepository implements cart.CartRepository using GORM.
type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository creates a new GORM-backed CartRepository.
func NewCartRepository(db *gorm.DB) cart.CartRepository {
	return &cartRepository{db: db}
}

// FindByUserID retrieves all cart items for a given user.
func (r *cartRepository) FindByUserID(ctx context.Context, userID string) ([]cart.Cart, error) {
	db := config.DBFromContext(ctx, r.db)
	var items []cart.Cart
	err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}

// FindByUserIDAfterCursor retrieves cart items for a user using cursor-based pagination.
// Items are ordered by ID ascending. When cursor is non-empty, only items with id > cursor are returned.
// The query fetches limit items.
func (r *cartRepository) FindByUserIDAfterCursor(ctx context.Context, userID, cursor string, limit int) ([]cart.Cart, error) {
	db := config.DBFromContext(ctx, r.db)

	query := db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id ASC").
		Limit(limit)

	if cursor != "" {
		query = query.Where("id > ?", cursor)
	}

	var items []cart.Cart
	err := query.Find(&items).Error
	return items, err
}

// FindByUserAndProduct retrieves a specific cart item by user and product.
// Returns nil if not found.
func (r *cartRepository) FindByUserAndProduct(ctx context.Context, userID, productID string) (*cart.Cart, error) {
	db := config.DBFromContext(ctx, r.db)
	var c cart.Cart
	err := db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&c).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// Save persists a cart item (create or update).
func (r *cartRepository) Save(ctx context.Context, c *cart.Cart) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(c).Error
}

// Delete removes a cart item.
func (r *cartRepository) Delete(ctx context.Context, c *cart.Cart) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Delete(c).Error
}
