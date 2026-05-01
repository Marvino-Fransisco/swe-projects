package command

import (
	"context"

	"inventory-service/internal/events"
)

// InventoryEventPublisher is a port for publishing inventory events.
// The concrete implementation is the RabbitMQ publisher adapter.
type InventoryEventPublisher interface {
	PublishStockReserved(ctx context.Context, event events.StockReservedEvent) error
	PublishStockRejected(ctx context.Context, event events.StockRejectedEvent) error
}
