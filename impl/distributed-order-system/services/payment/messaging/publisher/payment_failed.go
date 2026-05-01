package publisher

import (
	"context"
	"encoding/json"
	"log"

	"payment-service/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishPaymentFailed publishes a PaymentFailed event to the payments exchange.
// This method satisfies the command.PaymentEventPublisher interface.
func (p *Publisher) PublishPaymentFailed(ctx context.Context, event events.PaymentFailedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.channel.PublishWithContext(
		ctx,
		"payments",
		"payments.failed",
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

	log.Printf("Published PaymentFailed event for order %s", event.OrderID)
	return nil
}
