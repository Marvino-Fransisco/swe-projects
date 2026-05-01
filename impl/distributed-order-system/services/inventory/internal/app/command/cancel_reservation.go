package command

import (
	"context"
	"fmt"

	"inventory-service/internal/domain/inventory"
)

// CancelReservation is the command for cancelling reservations after payment fails.
type CancelReservation struct {
	OrderID string
}

// CancelReservationHandler processes the CancelReservation command.
// It marks reserved reservations as cancelled and restores stock to inventory.
type CancelReservationHandler struct {
	repo inventory.Repository
}

// NewCancelReservationHandler constructs a new handler with its dependencies.
func NewCancelReservationHandler(repo inventory.Repository) CancelReservationHandler {
	return CancelReservationHandler{repo: repo}
}

// Handle loads reserved reservations, marks them as cancelled, and restores stock.
func (h CancelReservationHandler) Handle(ctx context.Context, cmd CancelReservation) error {
	reservations, err := h.repo.FindReservationsByOrderID(ctx, cmd.OrderID, inventory.ReservationStatusReserved)
	if err != nil {
		return fmt.Errorf("failed to find reservations for order %s: %w", cmd.OrderID, err)
	}

	if len(reservations) == 0 {
		return nil
	}

	for i := range reservations {
		res := &reservations[i]

		// Cancel the reservation.
		res.Cancel()
		if err := h.repo.UpdateReservation(ctx, res); err != nil {
			return fmt.Errorf("failed to cancel reservation for order %s, product %s: %w", cmd.OrderID, res.ProductID(), err)
		}

		// Restore stock back to inventory.
		inventories, err := h.repo.FindByProductIDs(ctx, []string{res.ProductID()})
		if err != nil {
			return fmt.Errorf("failed to find inventory for product %s: %w", res.ProductID(), err)
		}
		if len(inventories) == 0 {
			return fmt.Errorf("product %s not found in inventory", res.ProductID())
		}

		inv := inventories[0]
		inv.RestoreStock(res.Quantity())
		if err := h.repo.UpdateInventory(ctx, &inv); err != nil {
			return fmt.Errorf("failed to restore stock for product %s: %w", res.ProductID(), err)
		}
	}

	return nil
}
