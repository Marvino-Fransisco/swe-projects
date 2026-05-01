package command

import (
	"context"
	"fmt"

	"order-service/internal/domain/order"
)

// UpdateOrderStatus is the command for changing an order's status.
type UpdateOrderStatus struct {
	OrderID string
	Status  order.OrderStatus
}

// UpdateOrderStatusHandler processes the UpdateOrderStatus command.
type UpdateOrderStatusHandler struct {
	repo order.Repository
}

// NewUpdateOrderStatusHandler constructs a new handler with its dependencies.
func NewUpdateOrderStatusHandler(repo order.Repository) UpdateOrderStatusHandler {
	return UpdateOrderStatusHandler{repo: repo}
}

// Handle loads the order, applies the state transition via domain logic, then persists.
func (h UpdateOrderStatusHandler) Handle(ctx context.Context, cmd UpdateOrderStatus) error {
	o, err := h.repo.GetByID(ctx, cmd.OrderID)
	if err != nil {
		return fmt.Errorf("failed to get order %s: %w", cmd.OrderID, err)
	}

	switch cmd.Status {
	case order.StatusConfirmed:
		if err := o.Confirm(); err != nil {
			return fmt.Errorf("failed to confirm order %s: %w", cmd.OrderID, err)
		}
	case order.StatusCancelled:
		if err := o.Cancel(); err != nil {
			return fmt.Errorf("failed to cancel order %s: %w", cmd.OrderID, err)
		}
	default:
		return fmt.Errorf("unsupported status transition to %q", cmd.Status)
	}

	if err := h.repo.Update(ctx, o); err != nil {
		return fmt.Errorf("failed to update order %s: %w", cmd.OrderID, err)
	}

	return nil
}
