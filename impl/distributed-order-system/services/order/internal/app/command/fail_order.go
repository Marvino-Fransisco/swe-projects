package command

import (
	"context"
	"fmt"

	"order-service/internal/domain/order"
)

// FailOrder is the command for marking an order as failed with a reason.
// This is a compensating transaction used when an upstream service rejects the order.
type FailOrder struct {
	OrderID       string
	FailureReason order.FailureReason
}

// FailOrderHandler processes the FailOrder command.
type FailOrderHandler struct {
	repo order.Repository
}

// NewFailOrderHandler constructs a new handler with its dependencies.
func NewFailOrderHandler(repo order.Repository) FailOrderHandler {
	return FailOrderHandler{repo: repo}
}

// Handle loads the order, applies the Fail transition via domain logic, then persists.
func (h FailOrderHandler) Handle(ctx context.Context, cmd FailOrder) error {
	o, err := h.repo.GetByID(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order %s: %w", cmd.OrderID, err)
	}

	if err := o.Fail(cmd.FailureReason); err != nil {
		return fmt.Errorf("failed to fail order %s: %w", cmd.OrderID, err)
	}

	if err := h.repo.Update(ctx, o); err != nil {
		return fmt.Errorf("failed to update order %s: %w", cmd.OrderID, err)
	}

	return nil
}
