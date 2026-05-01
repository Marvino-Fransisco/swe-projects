package dbrepository

import (
	"context"
	"fmt"

	"inventory-service/internal/domain/inventory"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GormInventoryRepository implements inventory.Repository using GORM + PostgreSQL.
type GormInventoryRepository struct {
	db *gorm.DB
}

// NewGormInventoryRepository creates a new repository instance.
func NewGormInventoryRepository(db *gorm.DB) *GormInventoryRepository {
	return &GormInventoryRepository{db: db}
}

// TxKey is the context key used to store the transaction-scoped *gorm.DB.
// The DBTransaction function in the command layer sets this value, and the
// repository layer reads it to transparently participate in the transaction.
var TxKey txKey

type txKey struct{}

// DBTransaction executes a function within a database transaction.
type DBTransaction func(ctx context.Context, fn func(ctx context.Context) error) error

// NewDBTransaction creates a DBTransaction using a GORM database connection.
// The transaction's *gorm.DB is stored in the context so that repository methods
// automatically participate in the transaction via db_(ctx).
func NewDBTransaction(db *gorm.DB) DBTransaction {
	return func(ctx context.Context, fn func(ctx context.Context) error) error {
		return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			txCtx := context.WithValue(ctx, TxKey, tx)
			return fn(txCtx)
		})
	}
}

// db returns the *gorm.DB to use for the given context.
// If a transaction has been started via DBTransaction, the transaction-scoped
// DB is extracted from context; otherwise the default DB is used.
func (r *GormInventoryRepository) db_(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db
}

// AutoMigrate runs GORM auto-migration for the persistence models.
func (r *GormInventoryRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&inventoryModel{}, &reservationModel{})
}

// FindByProductIDs loads inventories by a slice of product IDs.
// When called within a transaction (detected via context), it automatically
// adds SELECT FOR UPDATE to acquire row-level locks.
func (r *GormInventoryRepository) FindByProductIDs(ctx context.Context, productIDs []string) ([]inventory.Inventory, error) {
	query := r.db_(ctx).WithContext(ctx).Where("product_id IN ?", productIDs)

	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var models []inventoryModel
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find inventories by product IDs: %w", err)
	}

	result := make([]inventory.Inventory, 0, len(models))
	for _, m := range models {
		result = append(result, modelToInventory(&m))
	}
	return result, nil
}

// UpdateInventory persists changes to an existing inventory item.
func (r *GormInventoryRepository) UpdateInventory(ctx context.Context, inv *inventory.Inventory) error {
	result := r.db_(ctx).WithContext(ctx).Model(&inventoryModel{}).Where("id = ?", inv.ID()).Updates(map[string]interface{}{
		"stock":      inv.Stock(),
		"status":     inv.Status().String(),
		"updated_at": inv.UpdatedAt(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update inventory %d: %w", inv.ID(), result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("inventory not found: %d", inv.ID())
	}
	return nil
}

// SaveReservation persists a new reservation.
func (r *GormInventoryRepository) SaveReservation(ctx context.Context, res inventory.InventoryReservation) error {
	model := reservationToModel(&res)
	if err := r.db_(ctx).WithContext(ctx).Create(model).Error; err != nil {
		return fmt.Errorf("failed to save reservation: %w", err)
	}
	return nil
}

// FindReservationsByOrderID loads all reservations for a given order ID with the specified status.
func (r *GormInventoryRepository) FindReservationsByOrderID(ctx context.Context, orderID string, status inventory.ReservationStatus) ([]inventory.InventoryReservation, error) {
	var models []reservationModel
	if err := r.db_(ctx).WithContext(ctx).Where("order_id = ? AND status = ?", orderID, status.String()).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to find reservations for order %s: %w", orderID, err)
	}

	result := make([]inventory.InventoryReservation, 0, len(models))
	for _, m := range models {
		result = append(result, modelToReservation(&m))
	}
	return result, nil
}

// UpdateReservation persists status changes to an existing reservation.
func (r *GormInventoryRepository) UpdateReservation(ctx context.Context, res *inventory.InventoryReservation) error {
	result := r.db_(ctx).WithContext(ctx).Model(&reservationModel{}).Where("id = ?", res.ID()).Updates(map[string]interface{}{
		"status":     res.Status().String(),
		"updated_at": res.UpdatedAt(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update reservation %d: %w", res.ID(), result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("reservation not found: %d", res.ID())
	}
	return nil
}

// modelToInventory converts a GORM model to a domain Inventory.
func modelToInventory(m *inventoryModel) inventory.Inventory {
	return *inventory.ReconstructInventory(
		m.ID,
		m.ProductID,
		m.ProductName,
		m.Stock,
		m.Price,
		inventory.InventoryStatus(m.Status),
		m.CreatedAt,
		m.UpdatedAt,
	)
}

// reservationToModel converts a domain InventoryReservation to a GORM model.
func reservationToModel(r *inventory.InventoryReservation) *reservationModel {
	return &reservationModel{
		ID:        r.ID(),
		OrderID:   r.OrderID(),
		ProductID: r.ProductID(),
		Quantity:  r.Quantity(),
		Status:    r.Status().String(),
		CreatedAt: r.CreatedAt(),
		UpdatedAt: r.UpdatedAt(),
	}
}

// modelToReservation converts a GORM model to a domain InventoryReservation.
func modelToReservation(m *reservationModel) inventory.InventoryReservation {
	return inventory.ReconstructInventoryReservation(
		m.ID,
		m.OrderID,
		m.ProductID,
		m.Quantity,
		inventory.ReservationStatus(m.Status),
		m.CreatedAt,
		m.UpdatedAt,
	)
}
