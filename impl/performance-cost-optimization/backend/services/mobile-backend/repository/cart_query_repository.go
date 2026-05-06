package repository

import (
	"context"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/cart"
)

// CartQueryRepository defines the interface for cart query (read) operations
// that return more than one data item.
type CartQueryRepository interface {
	// FindByUserIDAfterCursor retrieves cart items for a user using cursor-based pagination.
	// Items are ordered by ID ascending. When cursor is non-empty, only items with id > cursor are returned.
	// limit is the maximum number of items to return.
	FindByUserIDAfterCursor(ctx context.Context, userID, cursor string, limit int) ([]cart.Cart, error)

	// FindAllByUserID retrieves ALL cart items for a given user.
	// Used for internal processing (e.g., building an order) where all items are needed.
	FindAllByUserID(ctx context.Context, userID string) ([]cart.Cart, error)
}

type cartQueryRepository struct {
	db *gorm.DB
}

// NewCartQueryRepository creates a new GORM-backed CartQueryRepository.
func NewCartQueryRepository(db *gorm.DB) CartQueryRepository {
	return &cartQueryRepository{db: db}
}

func (r *cartQueryRepository) FindByUserIDAfterCursor(ctx context.Context, userID, cursor string, limit int) ([]cart.Cart, error) {
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

func (r *cartQueryRepository) FindAllByUserID(ctx context.Context, userID string) ([]cart.Cart, error) {
	db := config.DBFromContext(ctx, r.db)
	var items []cart.Cart
	err := db.WithContext(ctx).Where("user_id = ?", userID).Find(&items).Error
	return items, err
}
