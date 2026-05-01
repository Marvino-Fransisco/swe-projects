package command

import (
	"context"
	"fmt"
	"time"

	"payment-service/internal/domain/payment"

	"github.com/google/uuid"
)

// CreatePaymentProduct holds the product data for a CreatePayment command.
type CreatePaymentProduct struct {
	ProductID string
	Quantity  int
	Price     float64
}

// CreatePayment is the command for creating a new payment (from an inventory reserved event).
type CreatePayment struct {
	OrderID  string
	Products []CreatePaymentProduct
}

// CreatePaymentResult is returned after successfully creating a payment.
type CreatePaymentResult struct {
	PaymentID  string
	OrderID    string
	TotalPrice float64
	Status     payment.PaymentStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// CreatePaymentHandler processes the CreatePayment command.
type CreatePaymentHandler struct {
	repo payment.Repository
}

// NewCreatePaymentHandler constructs a new handler with its dependencies.
func NewCreatePaymentHandler(repo payment.Repository) CreatePaymentHandler {
	return CreatePaymentHandler{repo: repo}
}

// Handle executes the CreatePayment command.
// It calculates the total price from products, creates the domain entity, and persists it.
func (h CreatePaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (CreatePaymentResult, error) {
	now := time.Now()

	var totalPrice float64
	for _, p := range cmd.Products {
		totalPrice += p.Price * float64(p.Quantity)
	}

	paymentID := uuid.New().String()
	p := payment.NewPayment(paymentID, cmd.OrderID, totalPrice, now)

	if err := h.repo.Save(ctx, p); err != nil {
		return CreatePaymentResult{}, fmt.Errorf("failed to save payment: %w", err)
	}

	return CreatePaymentResult{
		PaymentID:  p.ID(),
		OrderID:    p.OrderID(),
		TotalPrice: p.TotalPrice(),
		Status:     p.Status(),
		CreatedAt:  p.CreatedAt(),
		UpdatedAt:  p.UpdatedAt(),
	}, nil
}
