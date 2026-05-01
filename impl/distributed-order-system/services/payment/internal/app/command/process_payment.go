package command

import (
	"context"
	"fmt"
	"time"

	"payment-service/internal/domain/payment"
	"payment-service/internal/events"
)

// PaymentEventPublisher is a port for publishing payment events.
// The concrete implementation is the RabbitMQ publisher adapter.
type PaymentEventPublisher interface {
	PublishPaymentSucceeded(ctx context.Context, event events.PaymentSucceededEvent) error
	PublishPaymentFailed(ctx context.Context, event events.PaymentFailedEvent) error
}

// ProcessPayment is the command for processing (confirming/rejecting) a payment.
type ProcessPayment struct {
	PaymentID string
	Amount    float64
}

// ProcessPaymentResult is returned after processing a payment.
type ProcessPaymentResult struct {
	PaymentID  string
	OrderID    string
	TotalPrice float64
	Status     payment.PaymentStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ProcessPaymentHandler processes the ProcessPayment command.
type ProcessPaymentHandler struct {
	repo      payment.Repository
	publisher PaymentEventPublisher
}

// NewProcessPaymentHandler constructs a new handler with its dependencies.
func NewProcessPaymentHandler(repo payment.Repository, publisher PaymentEventPublisher) ProcessPaymentHandler {
	return ProcessPaymentHandler{
		repo:      repo,
		publisher: publisher,
	}
}

// Handle loads the payment, applies the Process transition via domain logic,
// persists the change, and publishes the corresponding event.
func (h ProcessPaymentHandler) Handle(ctx context.Context, cmd ProcessPayment) (ProcessPaymentResult, error) {
	p, err := h.repo.GetByID(ctx, cmd.PaymentID)
	if err != nil {
		return ProcessPaymentResult{}, fmt.Errorf("failed to get payment %s: %w", cmd.PaymentID, err)
	}

	if err := p.Process(cmd.Amount); err != nil {
		return ProcessPaymentResult{}, fmt.Errorf("failed to process payment %s: %w", cmd.PaymentID, err)
	}

	if err := h.repo.Update(ctx, p); err != nil {
		return ProcessPaymentResult{}, fmt.Errorf("failed to update payment %s: %w", cmd.PaymentID, err)
	}

	// Publish the corresponding event based on the resulting status.
	if p.Status() == payment.StatusSucceeded {
		if err := h.publisher.PublishPaymentSucceeded(ctx, events.PaymentSucceededEvent{
			OrderID: p.OrderID(),
		}); err != nil {
			// Payment is already persisted — log the failure but don't fail the command.
			fmt.Printf("Failed to publish PaymentSucceeded event for payment %s: %v\n", p.ID(), err)
		}
	} else {
		if err := h.publisher.PublishPaymentFailed(ctx, events.PaymentFailedEvent{
			OrderID: p.OrderID(),
		}); err != nil {
			fmt.Printf("Failed to publish PaymentFailed event for payment %s: %v\n", p.ID(), err)
		}
	}

	return ProcessPaymentResult{
		PaymentID:  p.ID(),
		OrderID:    p.OrderID(),
		TotalPrice: p.TotalPrice(),
		Status:     p.Status(),
		CreatedAt:  p.CreatedAt(),
		UpdatedAt:  p.UpdatedAt(),
	}, nil
}
