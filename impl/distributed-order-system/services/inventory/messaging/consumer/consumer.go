package consumer

import (
	"context"
	"fmt"
	"log"

	"inventory-service/internal/app"

	"github.com/redis/go-redis/v9"

	amqp "github.com/rabbitmq/amqp091-go"
	sharedRabbitMQ "shared/rabbitmq"
)

const maxRetryCount = 5

type Consumer struct {
	*sharedRabbitMQ.BaseConsumer
	app         *app.Application
	redisClient *redis.Client
}

func NewConsumer(amqpURL string, application *app.Application, redisClient *redis.Client) (*Consumer, error) {
	base, err := sharedRabbitMQ.NewBaseConsumer(amqpURL)
	if err != nil {
		return nil, err
	}

	c := &Consumer{
		BaseConsumer: base,
		app:          application,
		redisClient:  redisClient,
	}

	if err := c.setup(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to setup consumer: %w", err)
	}

	return c, nil
}

func (c *Consumer) setup() error {
	// --- Orders exchanges ---
	if err := c.DeclareExchangeWithDLX("orders"); err != nil {
		return err
	}

	// --- Orders main queue ---
	ordersQueue, err := c.DeclareQueue("inventories.orders", amqp.Table{
		"x-dead-letter-exchange":    "orders.dlx",
		"x-dead-letter-routing-key": "inventories.orders.retry",
	})
	if err != nil {
		return err
	}

	// --- Orders retry queue ---
	ordersRetryQueue, err := c.DeclareQueue("inventories.orders.retry", amqp.Table{
		"x-message-ttl":             int32(5000),
		"x-dead-letter-exchange":    "orders",
		"x-dead-letter-routing-key": "orders.created",
	})
	if err != nil {
		return err
	}

	// --- Orders DLQ ---
	ordersDLQ, err := c.DeclareQueue("inventories.orders.dlq", nil)
	if err != nil {
		return err
	}

	// --- Orders bindings ---
	if err := c.BindQueue(ordersQueue.Name, "orders.created", "orders"); err != nil {
		return err
	}

	if err := c.BindQueue(ordersRetryQueue.Name, "inventories.orders.retry", "orders.dlx"); err != nil {
		return err
	}

	if err := c.BindQueue(ordersDLQ.Name, "inventories.orders.failed", "orders.dlx"); err != nil {
		return err
	}

	// --- Payments exchanges ---
	if err := c.DeclareExchangeWithDLX("payments"); err != nil {
		return err
	}

	// --- Payments main queue ---
	paymentsQueue, err := c.DeclareQueue("inventories.payments", amqp.Table{
		"x-dead-letter-exchange":    "payments.dlx",
		"x-dead-letter-routing-key": "inventories.payments.retry",
	})
	if err != nil {
		return err
	}

	// --- Payments retry queue ---
	paymentsRetryQueue, err := c.DeclareQueue("inventories.payments.retry", amqp.Table{
		"x-message-ttl":             int32(5000),
		"x-dead-letter-exchange":    "payments",
		"x-dead-letter-routing-key": "payments.*",
	})
	if err != nil {
		return err
	}

	// --- Payments DLQ ---
	paymentsDLQ, err := c.DeclareQueue("inventories.payments.dlq", nil)
	if err != nil {
		return err
	}

	// --- Payments bindings ---
	if err := c.BindQueue(paymentsQueue.Name, "payments.*", "payments"); err != nil {
		return err
	}

	if err := c.BindQueue(paymentsRetryQueue.Name, "inventories.payments.retry", "payments.dlx"); err != nil {
		return err
	}

	if err := c.BindQueue(paymentsDLQ.Name, "inventories.payments.failed", "payments.dlx"); err != nil {
		return err
	}

	return nil
}

// removeClaimCheck deletes the claim check key from Redis.
// Errors are logged but not propagated — claim check cleanup is best-effort.
func (c *Consumer) removeClaimCheck(key, orderID string) {
	if err := c.redisClient.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to remove claim check from Redis for order %s (key=%s): %v", orderID, key, err)
	} else {
		log.Printf("Removed claim check from Redis for order %s (key=%s)", orderID, key)
	}
}
