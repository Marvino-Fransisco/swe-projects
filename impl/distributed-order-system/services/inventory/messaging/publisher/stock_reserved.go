package publisher

import (
	"context"
	"encoding/json"
	"log"

	"inventory-service/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishStockReserved publishes a StockReserved event to the inventories exchange.
// This method satisfies the command.InventoryEventPublisher interface.
func (p *Publisher) PublishStockReserved(ctx context.Context, event events.StockReservedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.Channel.PublishWithContext(
		ctx,
		"inventories",
		"inventories.reserved",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("Published StockReserved event for order %s", event.OrderID)
	return nil
}
