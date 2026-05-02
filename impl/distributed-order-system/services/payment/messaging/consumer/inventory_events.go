package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"payment-service/internal/app/command"
	"payment-service/internal/events"

	sharedRabbitMQ "shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Consumer) StartConsumingInventoryEvents() error {
	deliveries, err := c.Channel.Consume(
		"payments.inventories",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming inventory events: %w", err)
	}

	go c.handleInventoryDeliveries(deliveries)
	log.Println("Started consuming messages from payments.inventories queue")
	return nil
}

func (c *Consumer) handleInventoryDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		log.Printf("Received message with routing key: %s", d.RoutingKey)

		switch d.RoutingKey {
		case "inventories.reserved":
			c.handleInventoryReserved(d)
		default:
			log.Printf("Unknown routing key: %s, moving to DLQ", d.RoutingKey)
			if err := c.PublishToDLQ("inventories.dlx", "payments.inventories.failed", d.Body); err != nil {
				log.Printf("Failed to publish inventory message to DLQ: %v", err)
			}
			d.Ack(false)
		}
	}
}

func (c *Consumer) handleInventoryReserved(d amqp.Delivery) {
	var event events.StockReservedEvent
	if err := json.Unmarshal(d.Body, &event); err != nil {
		log.Printf("Failed to unmarshal StockReservedEvent: %v", err)
		if err := c.PublishToDLQ("inventories.dlx", "payments.inventories.failed", d.Body); err != nil {
			log.Printf("Failed to publish inventory message to DLQ: %v", err)
		}
		d.Ack(false)
		return
	}

	if event.OrderID == "" {
		log.Printf("Received inventory message with empty order_id")
		if err := c.PublishToDLQ("inventories.dlx", "payments.inventories.failed", d.Body); err != nil {
			log.Printf("Failed to publish inventory message to DLQ: %v", err)
		}
		d.Ack(false)
		return
	}

	products := make([]command.CreatePaymentProduct, 0, len(event.Products))
	for _, p := range event.Products {
		products = append(products, command.CreatePaymentProduct{
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
			Price:     p.Price,
		})
	}

	cmd := command.CreatePayment{
		OrderID:  event.OrderID,
		Products: products,
	}

	result, err := c.app.Commands.CreatePayment.Handle(context.Background(), cmd)
	if err != nil {
		log.Printf("Failed to create payment for order %s: %v", event.OrderID, err)

		retryCount := sharedRabbitMQ.GetRetryCount(d.Headers) + 1
		if retryCount >= sharedRabbitMQ.DefaultMaxRetries {
			log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", event.OrderID, sharedRabbitMQ.DefaultMaxRetries)
			if err := c.PublishToDLQ("inventories.dlx", "payments.inventories.failed", d.Body); err != nil {
				log.Printf("Failed to publish inventory message to DLQ: %v", err)
			}
		} else {
			log.Printf("Retrying message for order %s, attempt %d/%d", event.OrderID, retryCount, sharedRabbitMQ.DefaultMaxRetries)
			if err := c.PublishToRetry("inventories.dlx", "payments.inventories.retry", d.Body, retryCount); err != nil {
				log.Printf("Failed to publish inventory message to retry queue: %v", err)
			}
		}
		d.Ack(false)
		return
	}

	log.Printf("Created payment %s for order %s with total price %.2f", result.PaymentID, result.OrderID, result.TotalPrice)

	c.callWebhook(webhookPayload{
		PaymentID:  result.PaymentID,
		OrderID:    result.OrderID,
		TotalPrice: result.TotalPrice,
		Status:     string(result.Status),
		CreatedAt:  result.CreatedAt,
		UpdatedAt:  result.UpdatedAt,
	})

	d.Ack(false)
}

// webhookPayload is the payload sent to the api-gateway webhook.
type webhookPayload struct {
	PaymentID  string    `json:"paymentId"`
	OrderID    string    `json:"orderId"`
	TotalPrice float64   `json:"totalPrice"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
