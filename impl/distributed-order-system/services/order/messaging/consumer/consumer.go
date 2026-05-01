package consumer

import (
	"fmt"
	"log"

	"order-service/internal/app"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	app        *app.Application
}

func NewConsumer(amqpURL string, application *app.Application) (*Consumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	c := &Consumer{
		connection: conn,
		channel:    ch,
		app:        application,
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
	if err := c.channel.ExchangeDeclare(
		"inventories",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare inventories exchange: %w", err)
	}

	if err := c.channel.ExchangeDeclare(
		"inventories.dlx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare inventories.dlx exchange: %w", err)
	}

	// --- Inventories main queue ---
	inventoriesQueue, err := c.channel.QueueDeclare(
		"orders.inventories",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "inventories.dlx",
			"x-dead-letter-routing-key": "orders.inventories.retry",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare orders.inventories queue: %w", err)
	}

	// --- Inventories retry queue ---
	inventoriesRetryQueue, err := c.channel.QueueDeclare(
		"orders.inventories.retry",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":             int32(5000),
			"x-dead-letter-exchange":    "inventories",
			"x-dead-letter-routing-key": "inventories.rejected",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare orders.inventories.retry queue: %w", err)
	}

	// --- Inventories DLQ ---
	inventoriesDLQ, err := c.channel.QueueDeclare(
		"orders.inventories.dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare orders.inventories.dlq queue: %w", err)
	}

	// --- Inventories bindings ---
	if err := c.channel.QueueBind(
		inventoriesQueue.Name,
		"inventories.rejected",
		"inventories",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind orders.inventories to inventories exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		inventoriesRetryQueue.Name,
		"orders.inventories.retry",
		"inventories.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind orders.inventories.retry to inventories.dlx exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		inventoriesDLQ.Name,
		"orders.inventories.failed",
		"inventories.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind orders.inventories.dlq to inventories.dlx exchange: %w", err)
	}

	// --- Payments exchanges ---
	if err := c.channel.ExchangeDeclare(
		"payments",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare payments exchange: %w", err)
	}

	if err := c.channel.ExchangeDeclare(
		"payments.dlx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare payments.dlx exchange: %w", err)
	}

	// --- Payments main queue ---
	paymentsQueue, err := c.channel.QueueDeclare(
		"orders.payments",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "payments.dlx",
			"x-dead-letter-routing-key": "orders.payments.retry",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare orders.payments queue: %w", err)
	}

	// --- Payments retry queue ---
	paymentsRetryQueue, err := c.channel.QueueDeclare(
		"orders.payments.retry",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":             int32(5000),
			"x-dead-letter-exchange":    "payments",
			"x-dead-letter-routing-key": "payments.*",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare orders.payments.retry queue: %w", err)
	}

	// --- Payments DLQ ---
	paymentsDLQ, err := c.channel.QueueDeclare(
		"orders.payments.dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare orders.payments.dlq queue: %w", err)
	}

	// --- Payments bindings ---
	if err := c.channel.QueueBind(
		paymentsQueue.Name,
		"payments.*",
		"payments",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind orders.payments to payments exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		paymentsRetryQueue.Name,
		"orders.payments.retry",
		"payments.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind orders.payments.retry to payments.dlx exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		paymentsDLQ.Name,
		"orders.payments.failed",
		"payments.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind orders.payments.dlq to payments.dlx exchange: %w", err)
	}

	return nil
}

const maxRetries int32 = 5

func getRetryCount(headers amqp.Table) int32 {
	if headers == nil {
		return 0
	}
	count, ok := headers["x-retry-count"]
	if !ok {
		return 0
	}
	switch v := count.(type) {
	case int32:
		return v
	case int:
		return int32(v)
	case int64:
		return int32(v)
	default:
		return 0
	}
}

func (c *Consumer) publishToDLQ(exchange, routingKey string, body []byte) error {
	return c.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (c *Consumer) publishToRetry(exchange, routingKey string, body []byte, retryCount int32) error {
	return c.channel.Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers: amqp.Table{
				"x-retry-count": retryCount,
			},
		},
	)
}

func (c *Consumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.connection != nil {
		c.connection.Close()
	}
}
