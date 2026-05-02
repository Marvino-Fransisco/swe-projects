package command

import (
	"context"
	"fmt"
	"time"

	"inventory-service/internal/domain/inventory"
	"inventory-service/internal/events"

	sharedTx "shared/tx"
)

// OrderProduct holds the product data from an incoming order created event.
type OrderProduct struct {
	ProductID string
	Quantity  int
}

// ReserveStock is the command for reserving stock for an order.
type ReserveStock struct {
	OrderID   string
	Products  []OrderProduct
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReserveStockHandler processes the ReserveStock command.
// It validates stock availability, creates reservations, and publishes events.
type ReserveStockHandler struct {
	repo      inventory.Repository
	publisher InventoryEventPublisher
	dbTx      sharedTx.DBTransaction
}

// NewReserveStockHandler constructs a new handler with its dependencies.
func NewReserveStockHandler(repo inventory.Repository, publisher InventoryEventPublisher, dbTx sharedTx.DBTransaction) ReserveStockHandler {
	return ReserveStockHandler{
		repo:      repo,
		publisher: publisher,
		dbTx:      dbTx,
	}
}

// Handle executes the ReserveStock command within a database transaction.
// It uses SELECT FOR UPDATE to acquire row-level locks on inventory rows,
// preventing concurrent workers from reading stale stock values.
func (h ReserveStockHandler) Handle(ctx context.Context, cmd ReserveStock) error {
	if len(cmd.Products) == 0 {
		return fmt.Errorf("no products in order %s", cmd.OrderID)
	}

	productIDs := make([]string, len(cmd.Products))
	for i, p := range cmd.Products {
		productIDs[i] = p.ProductID
	}

	return h.dbTx(ctx, func(txCtx context.Context) error {
		inventories, err := h.repo.FindByProductIDs(txCtx, productIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch inventories for order %s: %w", cmd.OrderID, err)
		}

		if len(inventories) != len(productIDs) {
			_ = h.publisher.PublishStockRejected(ctx, events.StockRejectedEvent{
				OrderID:   cmd.OrderID,
				Products:  []events.InventoryProduct{},
				Status:    cmd.Status,
				CreatedAt: cmd.CreatedAt,
				UpdatedAt: cmd.UpdatedAt,
			})
			return fmt.Errorf("some products not found for order %s", cmd.OrderID)
		}

		inventoryMap := make(map[string]*inventory.Inventory, len(inventories))
		for i := range inventories {
			inventoryMap[inventories[i].ProductID()] = &inventories[i]
		}

		for _, p := range cmd.Products {
			inv := inventoryMap[p.ProductID]
			if !inv.HasSufficientStock(p.Quantity) {
				_ = h.publisher.PublishStockRejected(ctx, events.StockRejectedEvent{
					OrderID:   cmd.OrderID,
					Products:  []events.InventoryProduct{},
					Status:    cmd.Status,
					CreatedAt: cmd.CreatedAt,
					UpdatedAt: cmd.UpdatedAt,
				})
				return fmt.Errorf("insufficient stock for product %s: requested %d, available %d",
					inv.ProductName(), p.Quantity, inv.Stock())
			}
		}

		now := time.Now()
		for _, p := range cmd.Products {
			reservation := inventory.NewInventoryReservation(cmd.OrderID, p.ProductID, p.Quantity, now)
			if err := h.repo.SaveReservation(txCtx, reservation); err != nil {
				return fmt.Errorf("failed to save reservation for order %s, product %s: %w", cmd.OrderID, p.ProductID, err)
			}
		}

		for _, p := range cmd.Products {
			inv := inventoryMap[p.ProductID]
			inv.DeductStock(p.Quantity)
			if err := h.repo.UpdateInventory(txCtx, inv); err != nil {
				return fmt.Errorf("failed to deduct stock for product %s: %w", p.ProductID, err)
			}
		}

		eventProducts := make([]events.InventoryProduct, 0, len(cmd.Products))
		for _, p := range cmd.Products {
			inv := inventoryMap[p.ProductID]
			eventProducts = append(eventProducts, events.InventoryProduct{
				ProductID: inv.ProductID(),
				Quantity:  p.Quantity,
				Price:     inv.Price(),
			})
		}

		if err := h.publisher.PublishStockReserved(ctx, events.StockReservedEvent{
			OrderID:   cmd.OrderID,
			Products:  eventProducts,
			Status:    cmd.Status,
			CreatedAt: cmd.CreatedAt,
			UpdatedAt: cmd.UpdatedAt,
		}); err != nil {
			return fmt.Errorf("failed to publish StockReserved event for order %s: %w", cmd.OrderID, err)
		}

		return nil
	})
}
