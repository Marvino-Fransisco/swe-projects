package publisher

import (
	"context"
	"encoding/json"
	"log"

	"payment-service/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishPaymentSucceeded publishes a PaymentSucceeded event to the payments exchange.
// This method satisfies the command.PaymentEventPublisher interface.
func (p *Publisher) PublishPaymentSucceeded(ctx context.Context, event events.PaymentSucceededEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = p.Channel.PublishWithContext(
		ctx,
		"payments",
		"payments.succeeded",
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

	log.Printf("Published PaymentSucceeded event for order %s", event.OrderID)
	return nil
}
