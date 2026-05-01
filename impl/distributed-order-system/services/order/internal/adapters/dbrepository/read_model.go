package dbrepository

import (
	"context"

	"order-service/internal/app/query"

	"gorm.io/gorm"
)

// GormOrderReadModel implements query.OrderReadModel using GORM + PostgreSQL.
// The read model queries the same database but returns view DTOs instead of domain entities.
type GormOrderReadModel struct {
	db *gorm.DB
}

// NewGormOrderReadModel creates a new read model instance.
func NewGormOrderReadModel(db *gorm.DB) *GormOrderReadModel {
	return &GormOrderReadModel{db: db}
}

// GetOrderByID fetches a single order view by ID.
func (m *GormOrderReadModel) GetOrderByID(ctx context.Context, id string) (query.OrderView, error) {
	var model orderModel
	if err := m.db.WithContext(ctx).Preload("Products").Where("id = ?", id).First(&model).Error; err != nil {
		return query.OrderView{}, err
	}
	return modelToView(&model), nil
}

// ListOrders fetches all orders as view DTOs.
func (m *GormOrderReadModel) ListOrders(ctx context.Context) ([]query.OrderView, error) {
	var models []orderModel
	if err := m.db.WithContext(ctx).Preload("Products").Find(&models).Error; err != nil {
		return nil, err
	}

	views := make([]query.OrderView, 0, len(models))
	for i := range models {
		views = append(views, modelToView(&models[i]))
	}
	return views, nil
}

// modelToView converts a GORM model to a query OrderView (read-side DTO).
func modelToView(m *orderModel) query.OrderView {
	products := make([]query.ProductView, 0, len(m.Products))
	for _, p := range m.Products {
		products = append(products, query.ProductView{
			ID:        p.ID,
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}
	return query.OrderView{
		ID:            m.ID,
		Products:      products,
		Status:        m.Status,
		FailureReason: m.FailureReason,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
