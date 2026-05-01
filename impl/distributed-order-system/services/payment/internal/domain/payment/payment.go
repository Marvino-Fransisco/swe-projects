package payment

import (
	"fmt"
	"time"
)

// Payment is the aggregate root. It encapsulates all business rules for payment lifecycle.
type Payment struct {
	id         string
	orderID    string
	totalPrice float64
	status     PaymentStatus
	createdAt  time.Time
	updatedAt  time.Time
}

// NewPayment creates a new Payment in "pending" status.
func NewPayment(id string, orderID string, totalPrice float64, now time.Time) *Payment {
	return &Payment{
		id:         id,
		orderID:    orderID,
		totalPrice: totalPrice,
		status:     StatusPending,
		createdAt:  now,
		updatedAt:  now,
	}
}

// ReconstructPayment rebuilds a Payment from persistence (used by repository).
func ReconstructPayment(id string, orderID string, totalPrice float64, status PaymentStatus, createdAt, updatedAt time.Time) *Payment {
	return &Payment{
		id:         id,
		orderID:    orderID,
		totalPrice: totalPrice,
		status:     status,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Getters — external code cannot mutate state directly.

func (p *Payment) ID() string           { return p.id }
func (p *Payment) OrderID() string      { return p.orderID }
func (p *Payment) TotalPrice() float64  { return p.totalPrice }
func (p *Payment) Status() PaymentStatus { return p.status }
func (p *Payment) CreatedAt() time.Time { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time { return p.updatedAt }

// Process transitions the payment based on the provided amount.
// If amount >= totalPrice, the payment succeeds; otherwise it fails.
// Only pending payments can be processed.
func (p *Payment) Process(amount float64) error {
	if p.status != StatusPending {
		return fmt.Errorf("cannot process payment: current status is %q, expected %q", p.status, StatusPending)
	}
	if amount >= p.totalPrice {
		p.status = StatusSucceeded
	} else {
		p.status = StatusFailed
	}
	p.updatedAt = time.Now()
	return nil
}
