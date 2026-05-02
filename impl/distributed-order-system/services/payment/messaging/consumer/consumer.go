package consumer

import (
	"fmt"

	"payment-service/internal/app"

	sharedRabbitMQ "shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer wraps the shared BaseConsumer and handles payment-specific message routing.
type Consumer struct {
	*sharedRabbitMQ.BaseConsumer
	app *app.Application
}

// NewConsumer creates a new Consumer by dialing RabbitMQ, setting up the exchange/queue
// topology, and preparing to consume inventory events.
func NewConsumer(amqpURL string, application *app.Application) (*Consumer, error) {
	base, err := sharedRabbitMQ.NewBaseConsumer(amqpURL)
	if err != nil {
		return nil, err
	}

	c := &Consumer{
		BaseConsumer: base,
		app:          application,
	}

	if err := c.setup(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to setup consumer: %w", err)
	}

	return c, nil
}

func (c *Consumer) setup() error {
	// --- Inventories exchanges ---
	if err := c.DeclareExchangeWithDLX("inventories"); err != nil {
		return fmt.Errorf("failed to declare inventories exchange: %w", err)
	}

	// --- Inventories main queue ---
	inventoriesQueue, err := c.DeclareQueue("payments.inventories", amqp.Table{
		"x-dead-letter-exchange":    "inventories.dlx",
		"x-dead-letter-routing-key": "payments.inventories.retry",
	})
	if err != nil {
		return fmt.Errorf("failed to declare payments.inventories queue: %w", err)
	}

	// --- Inventories retry queue ---
	inventoriesRetryQueue, err := c.DeclareQueue("payments.inventories.retry", amqp.Table{
		"x-message-ttl":             int32(5000),
		"x-dead-letter-exchange":    "inventories",
		"x-dead-letter-routing-key": "inventories.reserved",
	})
	if err != nil {
		return fmt.Errorf("failed to declare payments.inventories.retry queue: %w", err)
	}

	// --- Inventories DLQ ---
	inventoriesDLQ, err := c.DeclareQueue("payments.inventories.dlq", nil)
	if err != nil {
		return fmt.Errorf("failed to declare payments.inventories.dlq queue: %w", err)
	}

	// --- Inventories bindings ---
	if err := c.BindQueue(inventoriesQueue.Name, "inventories.reserved", "inventories"); err != nil {
		return fmt.Errorf("failed to bind payments.inventories to inventories exchange: %w", err)
	}

	if err := c.BindQueue(inventoriesRetryQueue.Name, "payments.inventories.retry", "inventories.dlx"); err != nil {
		return fmt.Errorf("failed to bind payments.inventories.retry to inventories.dlx exchange: %w", err)
	}

	if err := c.BindQueue(inventoriesDLQ.Name, "payments.inventories.failed", "inventories.dlx"); err != nil {
		return fmt.Errorf("failed to bind payments.inventories.dlq to inventories.dlx exchange: %w", err)
	}

	return nil
}

// Close releases the underlying channel and connection via BaseConsumer.
func (c *Consumer) Close() {
	if c.BaseConsumer != nil {
		c.BaseConsumer.Close()
	}
}
