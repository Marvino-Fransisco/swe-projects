package command

import (
	"context"
	"fmt"
	"log"
	"time"

	"order-service/internal/domain/order"
	"order-service/internal/events"

	"github.com/google/uuid"
)

// OrderEventPublisher is a port for publishing order events.
// The concrete implementation is the RabbitMQ publisher adapter.
type OrderEventPublisher interface {
	PublishOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error
}

// CreateOrderProduct holds the raw product data for a new order command.
type CreateOrderProduct struct {
	ProductID string
	Quantity  int
}

// CreateOrder is the command for creating a new order.
type CreateOrder struct {
	OrderID  string
	Products []CreateOrderProduct
}

// CreateOrderHandler processes the CreateOrder command.
type CreateOrderHandler struct {
	repo             order.Repository
	publisher        OrderEventPublisher
	failOrderHandler FailOrderHandler
}

// NewCreateOrderHandler constructs a new handler with its dependencies.
func NewCreateOrderHandler(repo order.Repository, publisher OrderEventPublisher, failOrderHandler FailOrderHandler) CreateOrderHandler {
	return CreateOrderHandler{
		repo:             repo,
		publisher:        publisher,
		failOrderHandler: failOrderHandler,
	}
}

// Handle executes the CreateOrder command.
// It creates the domain entity, persists it, and publishes an event.
// If publishing fails, a compensating transaction marks the order as failed.
func (h CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrder) error {
	now := time.Now()

	products := make([]order.OrderProduct, 0, len(cmd.Products))
	for _, p := range cmd.Products {
		products = append(products, order.NewOrderProduct(
			uuid.New().String(),
			p.ProductID,
			p.Quantity,
			now,
		))
	}

	o := order.NewOrder(cmd.OrderID, products, now)

	if err := h.repo.Save(ctx, o); err != nil {
		return fmt.Errorf("failed to save order: %w", err)
	}

	eventProducts := make([]events.OrderProductEvent, 0, len(o.Products()))
	for _, p := range o.Products() {
		eventProducts = append(eventProducts, events.OrderProductEvent{
			ProductID: p.ProductID(),
			Quantity:  p.Quantity(),
		})
	}

	event := events.OrderCreatedEvent{
		ID:        o.ID(),
		Products:  eventProducts,
		Status:    o.Status().String(),
		CreatedAt: o.CreatedAt(),
		UpdatedAt: o.UpdatedAt(),
	}

	if err := h.publisher.PublishOrderCreated(ctx, event); err != nil {
		log.Printf("Failed to publish OrderCreated event for order %s: %v", o.ID(), err)

		// Compensating transaction: mark the order as failed.
		if failErr := h.failOrderHandler.Handle(ctx, FailOrder{
			OrderID:       o.ID(),
			FailureReason: order.FailureReasonPublishFail,
		}); failErr != nil {
			log.Printf("CRITICAL: Failed to compensate order %s after publish failure: %v", o.ID(), failErr)
			return fmt.Errorf("publish failed and compensating transaction also failed for order %s: %w", o.ID(), err)
		}

		log.Printf("Compensating transaction: order %s marked as failed (reason: %s)", o.ID(), order.FailureReasonPublishFail)
		return fmt.Errorf("failed to publish OrderCreated event for order %s: %w", o.ID(), err)
	}

	return nil
}
