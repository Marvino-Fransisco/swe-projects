package dbrepository

import (
	"context"
	"fmt"
	"math"

	"inventory-service/internal/app/query"

	"gorm.io/gorm"
)

// GormInventoryReadModel implements query.InventoryReadModel using GORM + PostgreSQL.
// The read model queries the same database but returns view DTOs instead of domain entities.
type GormInventoryReadModel struct {
	db *gorm.DB
}

// NewGormInventoryReadModel creates a new read model instance.
func NewGormInventoryReadModel(db *gorm.DB) *GormInventoryReadModel {
	return &GormInventoryReadModel{db: db}
}

// ListInventories fetches paginated inventories as view DTOs.
func (m *GormInventoryReadModel) ListInventories(ctx context.Context, page, limit int) (query.PaginationResult, error) {
	var total int64
	if err := m.db.WithContext(ctx).Model(&inventoryModel{}).Count(&total).Error; err != nil {
		return query.PaginationResult{}, fmt.Errorf("failed to count inventories: %w", err)
	}

	offset := (page - 1) * limit
	var models []inventoryModel
	if err := m.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return query.PaginationResult{}, fmt.Errorf("failed to fetch inventories: %w", err)
	}

	views := make([]query.InventoryView, 0, len(models))
	for _, model := range models {
		views = append(views, modelToView(&model))
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return query.PaginationResult{
		Data:       views,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// modelToView converts a GORM model to a query InventoryView (read-side DTO).
func modelToView(m *inventoryModel) query.InventoryView {
	return query.InventoryView{
		ID:          m.ID,
		ProductID:   m.ProductID,
		ProductName: m.ProductName,
		Stock:       m.Stock,
		Price:       m.Price,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
