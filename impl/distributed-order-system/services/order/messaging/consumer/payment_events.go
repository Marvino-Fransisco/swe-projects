package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"order-service/internal/app/command"
	"order-service/internal/domain/order"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Consumer) StartConsumingPaymentEvents() error {
	deliveries, err := c.channel.Consume(
		"orders.payments",
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
	log.Println("Started consuming messages from orders.payments queue")
	return nil
}

func (c *Consumer) handlePaymentDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		log.Printf("Received message with routing key: %s", d.RoutingKey)

		var msg PaymentMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("Failed to unmarshal payment message: %v", err)
			d.Ack(false)
			continue
		}

		if msg.OrderID == "" {
			log.Printf("Received payment message with empty order_id")
			d.Ack(false)
			continue
		}

		var status order.OrderStatus
		switch d.RoutingKey {
		case "payments.succeeded":
			status = order.StatusConfirmed
		case "payments.failed":
			status = order.StatusCancelled
		default:
			log.Printf("Unknown routing key: %s", d.RoutingKey)
			d.Ack(false)
			continue
		}

		if err := c.app.Commands.UpdateOrderStatus.Handle(context.Background(), command.UpdateOrderStatus{
			OrderID: msg.OrderID,
			Status:  status,
		}); err != nil {
			log.Printf("Failed to update order %s status to %s: %v", msg.OrderID, status, err)
			retryCount := getRetryCount(d.Headers) + 1
			if retryCount >= maxRetries {
				log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", msg.OrderID, maxRetries)
				if err := c.publishToDLQ("payments.dlx", "orders.payments.failed", d.Body); err != nil {
					log.Printf("Failed to publish payment message to DLQ: %v", err)
				}
			} else {
				log.Printf("Retrying message for order %s, attempt %d/%d", msg.OrderID, retryCount, maxRetries)
				if err := c.publishToRetry("payments.dlx", "orders.payments.retry", d.Body, retryCount); err != nil {
					log.Printf("Failed to publish payment message to retry queue: %v", err)
				}
			}
			d.Ack(false)
			continue
		}

		log.Printf("Updated order %s status to %s", msg.OrderID, status)
		d.Ack(false)
	}
}
