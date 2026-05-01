package inventory

import "context"

// Repository defines the persistence contract for Inventory and InventoryReservation.
// The write side (commands) uses this interface.
type Repository interface {
	// FindByProductIDs loads inventories by a slice of product IDs.
	// Automatically uses SELECT FOR UPDATE when called within a transaction context.
	FindByProductIDs(ctx context.Context, productIDs []string) ([]Inventory, error)

	// UpdateInventory persists changes to an existing inventory item (e.g. stock deduction).
	UpdateInventory(ctx context.Context, inventory *Inventory) error

	// SaveReservation persists a new reservation.
	SaveReservation(ctx context.Context, reservation InventoryReservation) error

	// FindReservationsByOrderID loads all reservations for a given order ID with the specified status.
	FindReservationsByOrderID(ctx context.Context, orderID string, status ReservationStatus) ([]InventoryReservation, error)

	// UpdateReservation persists status changes to an existing reservation.
	UpdateReservation(ctx context.Context, reservation *InventoryReservation) error
}
