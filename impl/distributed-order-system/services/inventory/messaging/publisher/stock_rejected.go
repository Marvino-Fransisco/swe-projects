package publisher

import (
	"context"
	"encoding/json"
	"log"

	"inventory-service/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishStockRejected publishes a StockRejected event to the inventories exchange.
// This method satisfies the command.InventoryEventPublisher interface.
func (p *Publisher) PublishStockRejected(ctx context.Context, event events.StockRejectedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		"inventories",
		"inventories.rejected",
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

	log.Printf("Published StockRejected event for order %s", event.OrderID)
	return nil
}
