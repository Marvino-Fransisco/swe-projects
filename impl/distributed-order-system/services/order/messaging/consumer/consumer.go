package consumer

import (
	"fmt"
	"log"

	"order-service/internal/app"

	sharedRabbitMQ "shared/rabbitmq"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	*sharedRabbitMQ.BaseConsumer
	app *app.Application
}

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

	log.Println("Consumer connected to RabbitMQ successfully")
	return c, nil
}

func (c *Consumer) setup() error {
	// --- Inventories exchanges ---
	if err := c.DeclareExchangeWithDLX("inventories"); err != nil {
		return err
	}

	// --- Inventories main queue ---
	inventoriesQueue, err := c.DeclareQueue("orders.inventories", amqp.Table{
		"x-dead-letter-exchange":    "inventories.dlx",
		"x-dead-letter-routing-key": "orders.inventories.retry",
	})
	if err != nil {
		return err
	}

	// --- Inventories retry queue ---
	inventoriesRetryQueue, err := c.DeclareQueue("orders.inventories.retry", amqp.Table{
		"x-message-ttl":             int32(5000),
		"x-dead-letter-exchange":    "inventories",
		"x-dead-letter-routing-key": "inventories.rejected",
	})
	if err != nil {
		return err
	}

	// --- Inventories DLQ ---
	inventoriesDLQ, err := c.DeclareQueue("orders.inventories.dlq", nil)
	if err != nil {
		return err
	}

	// --- Inventories bindings ---
	if err := c.BindQueue(inventoriesQueue.Name, "inventories.rejected", "inventories"); err != nil {
		return err
	}

	if err := c.BindQueue(inventoriesRetryQueue.Name, "orders.inventories.retry", "inventories.dlx"); err != nil {
		return err
	}

	if err := c.BindQueue(inventoriesDLQ.Name, "orders.inventories.failed", "inventories.dlx"); err != nil {
		return err
	}

	// --- Payments exchanges ---
	if err := c.DeclareExchangeWithDLX("payments"); err != nil {
		return err
	}

	// --- Payments main queue ---
	paymentsQueue, err := c.DeclareQueue("orders.payments", amqp.Table{
		"x-dead-letter-exchange":    "payments.dlx",
		"x-dead-letter-routing-key": "orders.payments.retry",
	})
	if err != nil {
		return err
	}

	// --- Payments retry queue ---
	paymentsRetryQueue, err := c.DeclareQueue("orders.payments.retry", amqp.Table{
		"x-message-ttl":             int32(5000),
		"x-dead-letter-exchange":    "payments",
		"x-dead-letter-routing-key": "payments.*",
	})
	if err != nil {
		return err
	}

	// --- Payments DLQ ---
	paymentsDLQ, err := c.DeclareQueue("orders.payments.dlq", nil)
	if err != nil {
		return err
	}

	// --- Payments bindings ---
	if err := c.BindQueue(paymentsQueue.Name, "payments.*", "payments"); err != nil {
		return err
	}

	if err := c.BindQueue(paymentsRetryQueue.Name, "orders.payments.retry", "payments.dlx"); err != nil {
		return err
	}

	if err := c.BindQueue(paymentsDLQ.Name, "orders.payments.failed", "payments.dlx"); err != nil {
		return err
	}

	return nil
}
