package repository

import (
	"context"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/order"
)

// OrderQueryRepository defines the interface for order read operations.
type OrderQueryRepository interface {
	// FindByUserPaginated retrieves a paginated list of orders for a user.
	// Returns the orders and the total count of matching records.
	FindByUserPaginated(ctx context.Context, userID string, page, pageSize int) ([]order.Order, int64, error)
}

type orderQueryRepository struct {
	db *gorm.DB
}

// NewOrderQueryRepository creates a new GORM-backed OrderQueryRepository.
func NewOrderQueryRepository(db *gorm.DB) OrderQueryRepository {
	return &orderQueryRepository{db: db}
}

func (r *orderQueryRepository) FindByUserPaginated(ctx context.Context, userID string, page, pageSize int) ([]order.Order, int64, error) {
	db := config.DBFromContext(ctx, r.db)

	var total int64
	if err := db.WithContext(ctx).Model(&order.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var orders []order.Order
	offset := (page - 1) * pageSize
	err := db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error

	return orders, total, err
}
