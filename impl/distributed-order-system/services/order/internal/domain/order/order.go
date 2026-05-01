package order

import (
	"fmt"
	"time"
)

// Order is the aggregate root. It encapsulates all business rules for order lifecycle.
type Order struct {
	id            string
	products      []OrderProduct
	status        OrderStatus
	failureReason FailureReason
	createdAt     time.Time
	updatedAt     time.Time
}

// NewOrder creates a new Order in "pending" status.
func NewOrder(id string, products []OrderProduct, now time.Time) *Order {
	return &Order{
		id:            id,
		products:      products,
		status:        StatusPending,
		failureReason: FailureReasonNone,
		createdAt:     now,
		updatedAt:     now,
	}
}

// ReconstructOrder rebuilds an Order from persistence (used by repository).
func ReconstructOrder(id string, products []OrderProduct, status OrderStatus, failureReason FailureReason, createdAt, updatedAt time.Time) *Order {
	return &Order{
		id:            id,
		products:      products,
		status:        status,
		failureReason: failureReason,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

// Getters — external code cannot mutate state directly.

func (o *Order) ID() string                     { return o.id }
func (o *Order) Products() []OrderProduct        { return o.products }
func (o *Order) Status() OrderStatus             { return o.status }
func (o *Order) FailureReason() FailureReason    { return o.failureReason }
func (o *Order) CreatedAt() time.Time            { return o.createdAt }
func (o *Order) UpdatedAt() time.Time            { return o.updatedAt }

// Confirm transitions the order from "pending" to "confirmed".
// This is called when payment succeeds.
func (o *Order) Confirm() error {
	if o.status != StatusPending {
		return fmt.Errorf("cannot confirm order: current status is %q, expected %q", o.status, StatusPending)
	}
	o.status = StatusConfirmed
	o.updatedAt = time.Now()
	return nil
}

// Cancel transitions the order to "cancelled".
// This is called when payment fails or inventory rejects.
func (o *Order) Cancel() error {
	if o.status == StatusCancelled {
		return fmt.Errorf("order is already cancelled")
	}
	if o.status == StatusConfirmed {
		return fmt.Errorf("cannot cancel confirmed order")
	}
	o.status = StatusCancelled
	o.updatedAt = time.Now()
	return nil
}

// Fail transitions the order to "failed" with a given reason.
// This is a compensating action used when an upstream service rejects the order.
func (o *Order) Fail(reason FailureReason) error {
	if o.status == StatusFailed {
		return fmt.Errorf("order is already failed")
	}
	if o.status == StatusConfirmed {
		return fmt.Errorf("cannot fail confirmed order")
	}
	o.status = StatusFailed
	o.failureReason = reason
	o.updatedAt = time.Now()
	return nil
}
