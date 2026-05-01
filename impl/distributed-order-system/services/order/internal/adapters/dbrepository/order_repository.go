package dbrepository

import (
	"context"
	"fmt"

	"order-service/internal/domain/order"

	"gorm.io/gorm"
)

// GormOrderRepository implements order.Repository using GORM + PostgreSQL.
type GormOrderRepository struct {
	db *gorm.DB
}

// NewGormOrderRepository creates a new repository instance.
func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

// AutoMigrate runs GORM auto-migration for the persistence models.
func (r *GormOrderRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&orderModel{}, &orderProductModel{})
}

// Save persists a new order (INSERT with associated products).
func (r *GormOrderRepository) Save(ctx context.Context, o *order.Order) error {
	model := orderToModel(o)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}
	return nil
}

// GetByID loads an order by ID with preloaded products.
func (r *GormOrderRepository) GetByID(ctx context.Context, id string) (*order.Order, error) {
	var model orderModel
	if err := r.db.WithContext(ctx).Preload("Products").Where("id = ?", id).First(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to get order %s: %w", id, err)
	}
	return modelToOrder(&model), nil
}

// Update persists status changes to an existing order.
func (r *GormOrderRepository) Update(ctx context.Context, o *order.Order) error {
	result := r.db.WithContext(ctx).Model(&orderModel{}).Where("id = ?", o.ID()).Updates(map[string]interface{}{
		"status":         o.Status().String(),
		"failure_reason": o.FailureReason().String(),
		"updated_at":     o.UpdatedAt(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update order %s: %w", o.ID(), result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order not found: %s", o.ID())
	}
	return nil
}

// orderToModel converts a domain Order to a GORM persistence model.
func orderToModel(o *order.Order) *orderModel {
	products := make([]orderProductModel, 0, len(o.Products()))
	for _, p := range o.Products() {
		products = append(products, orderProductModel{
			ID:        p.ID(),
			OrderID:   o.ID(),
			ProductID: p.ProductID(),
			Quantity:  p.Quantity(),
			CreatedAt: p.CreatedAt(),
			UpdatedAt: p.UpdatedAt(),
		})
	}
	return &orderModel{
		ID:            o.ID(),
		Products:      products,
		Status:        o.Status().String(),
		FailureReason: o.FailureReason().String(),
		CreatedAt:     o.CreatedAt(),
		UpdatedAt:     o.UpdatedAt(),
	}
}

// modelToOrder converts a GORM persistence model to a domain Order.
func modelToOrder(m *orderModel) *order.Order {
	products := make([]order.OrderProduct, 0, len(m.Products))
	for _, p := range m.Products {
		products = append(products, order.ReconstructOrderProduct(
			p.ID,
			p.OrderID,
			p.ProductID,
			p.Quantity,
			p.CreatedAt,
			p.UpdatedAt,
		))
	}
	return order.ReconstructOrder(
		m.ID,
		products,
		order.OrderStatus(m.Status),
		order.FailureReason(m.FailureReason),
		m.CreatedAt,
		m.UpdatedAt,
	)
}
