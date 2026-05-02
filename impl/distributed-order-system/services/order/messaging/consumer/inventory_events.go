package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"order-service/internal/app/command"
	"order-service/internal/domain/order"

	sharedRabbitMQ "shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webhookURL = "http://api-gateway_devcontainer-app-1:8080/api/webhooks"

func (c *Consumer) StartConsumingInventoryEvents() error {
	deliveries, err := c.Channel.Consume(
		"orders.inventories",
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
	log.Println("Started consuming messages from orders.inventories queue")
	return nil
}

func (c *Consumer) handleInventoryDeliveries(deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		log.Printf("Received inventory message with routing key: %s", d.RoutingKey)

		var msg InventoryMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			log.Printf("Failed to unmarshal inventory message: %v", err)
			if err := c.PublishToDLQ("inventories.dlx", "orders.inventories.failed", d.Body); err != nil {
				log.Printf("Failed to publish inventory message to DLQ: %v", err)
			}
			d.Ack(false)
			continue
		}

		if msg.OrderID == "" {
			log.Printf("Received inventory message with empty order_id")
			if err := c.PublishToDLQ("inventories.dlx", "orders.inventories.failed", d.Body); err != nil {
				log.Printf("Failed to publish inventory message to DLQ: %v", err)
			}
			d.Ack(false)
			continue
		}

		switch d.RoutingKey {
		case "inventories.rejected":
			if err := c.app.Commands.FailOrder.Handle(context.Background(), command.FailOrder{
				OrderID:       msg.OrderID,
				FailureReason: order.FailureReasonInsufficientInventory,
			}); err != nil {
				log.Printf("Failed to fail order %s: %v", msg.OrderID, err)
				retryCount := sharedRabbitMQ.GetRetryCount(d.Headers) + 1
				if retryCount >= sharedRabbitMQ.DefaultMaxRetries {
					log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", msg.OrderID, sharedRabbitMQ.DefaultMaxRetries)
					if err := c.PublishToDLQ("inventories.dlx", "orders.inventories.failed", d.Body); err != nil {
						log.Printf("Failed to publish inventory message to DLQ: %v", err)
					}
				} else {
					log.Printf("Retrying message for order %s, attempt %d/%d", msg.OrderID, retryCount, sharedRabbitMQ.DefaultMaxRetries)
					if err := c.PublishToRetry("inventories.dlx", "orders.inventories.retry", d.Body, retryCount); err != nil {
						log.Printf("Failed to publish inventory message to retry queue: %v", err)
					}
				}
				d.Ack(false)
				continue
			}

			log.Printf("Failed order %s due to insufficient inventory", msg.OrderID)

			if err := c.notifyStockRejected(msg.OrderID); err != nil {
				log.Printf("Failed to notify webhook for order %s: %v", msg.OrderID, err)
			}

		default:
			if err := c.app.Commands.UpdateOrderStatus.Handle(context.Background(), command.UpdateOrderStatus{
				OrderID: msg.OrderID,
				Status:  order.StatusCancelled,
			}); err != nil {
				log.Printf("Failed to update order %s status to cancelled: %v", msg.OrderID, err)
				retryCount := sharedRabbitMQ.GetRetryCount(d.Headers) + 1
				if retryCount >= sharedRabbitMQ.DefaultMaxRetries {
					log.Printf("Message for order %s exceeded max retries (%d), sending to DLQ", msg.OrderID, sharedRabbitMQ.DefaultMaxRetries)
					if err := c.PublishToDLQ("inventories.dlx", "orders.inventories.failed", d.Body); err != nil {
						log.Printf("Failed to publish inventory message to DLQ: %v", err)
					}
				} else {
					log.Printf("Retrying message for order %s, attempt %d/%d", msg.OrderID, retryCount, sharedRabbitMQ.DefaultMaxRetries)
					if err := c.PublishToRetry("inventories.dlx", "orders.inventories.retry", d.Body, retryCount); err != nil {
						log.Printf("Failed to publish inventory message to retry queue: %v", err)
					}
				}
				d.Ack(false)
				continue
			}

			log.Printf("Updated order %s status to cancelled (inventory rejected)", msg.OrderID)

			if err := c.notifyStockRejected(msg.OrderID); err != nil {
				log.Printf("Failed to notify webhook for order %s: %v", msg.OrderID, err)
			}
		}

		d.Ack(false)
	}
}

func (c *Consumer) notifyStockRejected(orderID string) error {
	payload := map[string]string{
		"orderId": orderID,
		"event":   "stock_rejected",
		"message": fmt.Sprintf("Stock insufficient for order %s", orderID),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to call webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned non-200 status: %d", resp.StatusCode)
	}

	log.Printf("Webhook notified for order %s: stock rejected", orderID)
	return nil
}
