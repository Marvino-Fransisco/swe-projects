package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"inventory-service/internal/app/command"

	sharedRabbitMQ "shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (c *Consumer) StartConsumingOrderEvents() error {
	deliveries, err := c.Channel.Consume(
		"inventories.orders",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming order events: %w", err)
	}

	workerCount := 3
	for i := 1; i <= workerCount; i++ {
		go c.handleOrderDeliveries(deliveries)
	}
	log.Printf("Started consuming messages from inventories.orders queue with %d workers", workerCount)
	return nil
}

func (c *Consumer) handleOrderDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		log.Printf("Received message with routing key: %s", d.RoutingKey)

		// Step 1: Unmarshal the claim check message.
		var claimCheck ClaimCheckMessage
		if err := json.Unmarshal(d.Body, &claimCheck); err != nil {
			log.Printf("Failed to unmarshal claim check message: %v", err)
			if err := c.PublishToDLQ("orders.dlx", "inventories.orders.failed", d.Body); err != nil {
				log.Printf("Failed to publish order message to DLQ: %v", err)
			}
			d.Ack(false)
			continue
		}

		if claimCheck.OrderID == "" || claimCheck.ClaimCheckKey == "" {
			log.Printf("Received claim check message with empty orderId or claimCheckKey")
			if err := c.PublishToDLQ("orders.dlx", "inventories.orders.failed", d.Body); err != nil {
				log.Printf("Failed to publish order message to DLQ: %v", err)
			}
			d.Ack(false)
			continue
		}

		// Step 2: Fetch the full payload from Redis using the claim check key.
		payloadBytes, err := c.redisClient.Get(context.Background(), claimCheck.ClaimCheckKey).Bytes()
		if err != nil {
			log.Printf("Failed to fetch payload from Redis for order %s (key=%s): %v", claimCheck.OrderID, claimCheck.ClaimCheckKey, err)
			retryCount := sharedRabbitMQ.GetRetryCount(d.Headers) + 1
			if retryCount >= maxRetryCount {
				log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", claimCheck.OrderID, maxRetryCount)
				if err := c.PublishToDLQ("orders.dlx", "inventories.orders.failed", d.Body); err != nil {
					log.Printf("Failed to publish order message to DLQ: %v", err)
				}
				c.removeClaimCheck(claimCheck.ClaimCheckKey, claimCheck.OrderID)
			} else {
				log.Printf("Retrying message for order %s, attempt %d/%d", claimCheck.OrderID, retryCount, maxRetryCount)
				if err := c.PublishToRetry("orders.dlx", "inventories.orders.retry", d.Body, retryCount); err != nil {
					log.Printf("Failed to publish order message to retry queue: %v", err)
				}
			}
			d.Ack(false)
			continue
		}

		// Step 3: Unmarshal the full payload.
		var msg OrderCreatedPayload
		if err := json.Unmarshal(payloadBytes, &msg); err != nil {
			log.Printf("Failed to unmarshal order payload from Redis for order %s: %v", claimCheck.OrderID, err)
			if err := c.PublishToDLQ("orders.dlx", "inventories.orders.failed", d.Body); err != nil {
				log.Printf("Failed to publish order message to DLQ: %v", err)
			}
			c.removeClaimCheck(claimCheck.ClaimCheckKey, claimCheck.OrderID)
			d.Ack(false)
			continue
		}

		if msg.ID == "" || len(msg.Products) == 0 {
			log.Printf("Received order payload with empty ID or no products for order %s", claimCheck.OrderID)
			if err := c.PublishToDLQ("orders.dlx", "inventories.orders.failed", d.Body); err != nil {
				log.Printf("Failed to publish order message to DLQ: %v", err)
			}
			c.removeClaimCheck(claimCheck.ClaimCheckKey, claimCheck.OrderID)
			d.Ack(false)
			continue
		}

		products := make([]command.OrderProduct, 0, len(msg.Products))
		for _, p := range msg.Products {
			products = append(products, command.OrderProduct{
				ProductID: p.ProductID,
				Quantity:  p.Quantity,
			})
		}

		cmd := command.ReserveStock{
			OrderID:   msg.ID,
			Products:  products,
			Status:    msg.Status,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := c.app.Commands.ReserveStock.Handle(context.Background(), cmd); err != nil {
			log.Printf("Failed to reserve stock for order %s: %v", msg.ID, err)
			if err := c.PublishToDLQ("orders.dlx", "inventories.orders.failed", d.Body); err != nil {
				log.Printf("Failed to publish order message to DLQ: %v", err)
			}
			c.removeClaimCheck(claimCheck.ClaimCheckKey, claimCheck.OrderID)
			d.Ack(false)
			continue
		}

		log.Printf("Reserved stock for order %s (%d products)", msg.ID, len(msg.Products))
		c.removeClaimCheck(claimCheck.ClaimCheckKey, claimCheck.OrderID)
		d.Ack(false)
	}
}
