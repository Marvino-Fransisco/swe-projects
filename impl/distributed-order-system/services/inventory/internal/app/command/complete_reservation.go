package command

import (
	"context"
	"fmt"

	"inventory-service/internal/domain/inventory"
)

// CompleteReservation is the command for completing reservations after payment succeeds.
type CompleteReservation struct {
	OrderID string
}

// CompleteReservationHandler processes the CompleteReservation command.
// It marks reservations as completed (stock was already deducted during reservation).
type CompleteReservationHandler struct {
	repo inventory.Repository
}

// NewCompleteReservationHandler constructs a new handler with its dependencies.
func NewCompleteReservationHandler(repo inventory.Repository) CompleteReservationHandler {
	return CompleteReservationHandler{repo: repo}
}

// Handle loads reserved reservations and marks them as completed.
func (h CompleteReservationHandler) Handle(ctx context.Context, cmd CompleteReservation) error {
	reservations, err := h.repo.FindReservationsByOrderID(ctx, cmd.OrderID, inventory.ReservationStatusReserved)
	if err != nil {
		return fmt.Errorf("failed to find reservations for order %s: %w", cmd.OrderID, err)
	}

	if len(reservations) == 0 {
		return nil
	}

	for i := range reservations {
		res := &reservations[i]

		// Complete the reservation.
		res.Complete()
		if err := h.repo.UpdateReservation(ctx, res); err != nil {
			return fmt.Errorf("failed to complete reservation for order %s, product %s: %w", cmd.OrderID, res.ProductID(), err)
		}
	}

	return nil
}
