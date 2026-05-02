package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"inventory-service/internal/app/command"

	sharedRabbitMQ "shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Consumer) StartConsumingPaymentEvents() error {
	deliveries, err := c.Channel.Consume(
		"inventories.payments",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming payment events: %w", err)
	}

	go c.handlePaymentDeliveries(deliveries)
	log.Println("Started consuming messages from inventories.payments queue")
	return nil
}

func (c *Consumer) handlePaymentDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		routingKey := d.RoutingKey
		log.Printf("Received payment message with routing key: %s", routingKey)

		var msg PaymentMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("Failed to unmarshal payment message: %v", err)
			if err := c.PublishToDLQ("payments.dlx", "inventories.payments.failed", d.Body); err != nil {
				log.Printf("Failed to publish payment message to DLQ: %v", err)
			}
			d.Ack(false)
			continue
		}

		if msg.OrderID == "" {
			log.Printf("Received payment message with empty order_id")
			if err := c.PublishToDLQ("payments.dlx", "inventories.payments.failed", d.Body); err != nil {
				log.Printf("Failed to publish payment message to DLQ: %v", err)
			}
			d.Ack(false)
			continue
		}

		switch routingKey {
		case "payments.succeeded":
			if err := c.app.Commands.CompleteReservation.Handle(context.Background(), command.CompleteReservation{
				OrderID: msg.OrderID,
			}); err != nil {
				log.Printf("Failed to complete reservation for order %s: %v", msg.OrderID, err)
				retryCount := sharedRabbitMQ.GetRetryCount(d.Headers) + 1
				if retryCount >= maxRetryCount {
					log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", msg.OrderID, maxRetryCount)
					if err := c.PublishToDLQ("payments.dlx", "inventories.payments.failed", d.Body); err != nil {
						log.Printf("Failed to publish payment message to DLQ: %v", err)
					}
				} else {
					log.Printf("Retrying message for order %s, attempt %d/%d", msg.OrderID, retryCount, maxRetryCount)
					if err := c.PublishToRetry("payments.dlx", "inventories.payments.retry", d.Body, retryCount); err != nil {
						log.Printf("Failed to publish payment message to retry queue: %v", err)
					}
				}
				d.Ack(false)
				continue
			}

			log.Printf("Completed reservations and deducted stock for order %s", msg.OrderID)

		case "payments.failed":
			if err := c.app.Commands.CancelReservation.Handle(context.Background(), command.CancelReservation{
				OrderID: msg.OrderID,
			}); err != nil {
				log.Printf("Failed to cancel reservation for order %s: %v", msg.OrderID, err)
				retryCount := sharedRabbitMQ.GetRetryCount(d.Headers) + 1
				if retryCount >= maxRetryCount {
					log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", msg.OrderID, maxRetryCount)
					if err := c.PublishToDLQ("payments.dlx", "inventories.payments.failed", d.Body); err != nil {
						log.Printf("Failed to publish payment message to DLQ: %v", err)
					}
				} else {
					log.Printf("Retrying message for order %s, attempt %d/%d", msg.OrderID, retryCount, maxRetryCount)
					if err := c.PublishToRetry("payments.dlx", "inventories.payments.retry", d.Body, retryCount); err != nil {
						log.Printf("Failed to publish payment message to retry queue: %v", err)
					}
				}
				d.Ack(false)
				continue
			}

			log.Printf("Cancelled reservations for order %s", msg.OrderID)

		default:
			log.Printf("Unknown payment routing key: %s", routingKey)
			if err := c.PublishToDLQ("payments.dlx", "inventories.payments.failed", d.Body); err != nil {
				log.Printf("Failed to publish payment message to DLQ: %v", err)
			}
		}

		d.Ack(false)
	}
}
