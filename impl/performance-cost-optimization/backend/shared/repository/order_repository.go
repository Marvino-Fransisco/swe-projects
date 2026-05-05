package repository

import (
	"context"

	"gorm.io/gorm"

	"shared/config"
	"shared/domain/order"
)

// orderRepository implements order.OrderRepository using GORM.
type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new GORM-backed OrderRepository.
func NewOrderRepository(db *gorm.DB) order.OrderRepository {
	return &orderRepository{db: db}
}

// Save persists a new order and its order details within a transaction.
func (r *orderRepository) Save(ctx context.Context, o *order.Order, details []order.OrderDetail) error {
	db := config.DBFromContext(ctx, r.db)

	if err := db.WithContext(ctx).Create(o).Error; err != nil {
		return err
	}

	if len(details) > 0 {
		if err := db.WithContext(ctx).Create(&details).Error; err != nil {
			return err
		}
	}

	return nil
}

// FindByID retrieves an order by its ID.
// Returns nil if not found.
func (r *orderRepository) FindByID(ctx context.Context, id string) (*order.Order, error) {
	db := config.DBFromContext(ctx, r.db)
	var o order.Order
	err := db.WithContext(ctx).Where("id = ?", id).First(&o).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// FindByIDAndUser retrieves an order by its ID filtered by userID.
// Returns nil if not found or does not belong to the user.
func (r *orderRepository) FindByIDAndUser(ctx context.Context, id, userID string) (*order.Order, error) {
	db := config.DBFromContext(ctx, r.db)
	var o order.Order
	err := db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&o).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

// FindDetailsByOrderID retrieves all order details for a given order.
func (r *orderRepository) FindDetailsByOrderID(ctx context.Context, orderID string) ([]order.OrderDetail, error) {
	db := config.DBFromContext(ctx, r.db)
	var details []order.OrderDetail
	err := db.WithContext(ctx).Where("order_id = ?", orderID).Find(&details).Error
	return details, err
}

// FindByUser retrieves a paginated list of orders for a user.
// Returns the orders and the total count of matching records.
func (r *orderRepository) FindByUser(ctx context.Context, userID string, page, pageSize int) ([]order.Order, int64, error) {
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

// Update persists changes to an existing order.
func (r *orderRepository) Update(ctx context.Context, o *order.Order) error {
	db := config.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Save(o).Error
}
