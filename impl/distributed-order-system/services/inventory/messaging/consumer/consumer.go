package consumer

import (
	"context"
	"fmt"
	"log"

	"inventory-service/internal/app"

	"github.com/redis/go-redis/v9"

	amqp "github.com/rabbitmq/amqp091-go"
)

const maxRetryCount = 5

type Consumer struct {
	connection  *amqp.Connection
	channel     *amqp.Channel
	app         *app.Application
	redisClient *redis.Client
}

func NewConsumer(amqpURL string, application *app.Application, redisClient *redis.Client) (*Consumer, error) {
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
		connection:  conn,
		channel:     ch,
		app:         application,
		redisClient: redisClient,
	}

	if err := c.setup(); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to setup consumer: %w", err)
	}

	log.Println("Consumer connected to RabbitMQ successfully")
	return c, nil
}

func (c *Consumer) setup() error {
	// --- Orders exchanges ---
	if err := c.channel.ExchangeDeclare(
		"orders",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare orders exchange: %w", err)
	}

	if err := c.channel.ExchangeDeclare(
		"orders.dlx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare orders.dlx exchange: %w", err)
	}

	// --- Orders main queue ---
	ordersQueue, err := c.channel.QueueDeclare(
		"inventories.orders",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "orders.dlx",
			"x-dead-letter-routing-key": "inventories.orders.retry",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare inventories.orders queue: %w", err)
	}

	// --- Orders retry queue ---
	ordersRetryQueue, err := c.channel.QueueDeclare(
		"inventories.orders.retry",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl":             int32(5000),
			"x-dead-letter-exchange":    "orders",
			"x-dead-letter-routing-key": "orders.created",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare inventories.orders.retry queue: %w", err)
	}

	// --- Orders DLQ ---
	ordersDLQ, err := c.channel.QueueDeclare(
		"inventories.orders.dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare inventories.orders.dlq queue: %w", err)
	}

	// --- Orders bindings ---
	if err := c.channel.QueueBind(
		ordersQueue.Name,
		"orders.created",
		"orders",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind inventories.orders to orders exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		ordersRetryQueue.Name,
		"inventories.orders.retry",
		"orders.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind inventories.orders.retry to orders.dlx exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		ordersDLQ.Name,
		"inventories.orders.failed",
		"orders.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind inventories.orders.dlq to orders.dlx exchange: %w", err)
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
		"inventories.payments",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "payments.dlx",
			"x-dead-letter-routing-key": "inventories.payments.retry",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare inventories.payments queue: %w", err)
	}

	// --- Payments retry queue ---
	paymentsRetryQueue, err := c.channel.QueueDeclare(
		"inventories.payments.retry",
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
		return fmt.Errorf("failed to declare inventories.payments.retry queue: %w", err)
	}

	// --- Payments DLQ ---
	paymentsDLQ, err := c.channel.QueueDeclare(
		"inventories.payments.dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to declare inventories.payments.dlq queue: %w", err)
	}

	// --- Payments bindings ---
	if err := c.channel.QueueBind(
		paymentsQueue.Name,
		"payments.*",
		"payments",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind inventories.payments to payments exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		paymentsRetryQueue.Name,
		"inventories.payments.retry",
		"payments.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind inventories.payments.retry to payments.dlx exchange: %w", err)
	}

	if err := c.channel.QueueBind(
		paymentsDLQ.Name,
		"inventories.payments.failed",
		"payments.dlx",
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind inventories.payments.dlq to payments.dlx exchange: %w", err)
	}

	return nil
}

// getRetryCount extracts the retry count from the x-retry-count header.
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

// removeClaimCheck deletes the claim check key from Redis.
// Errors are logged but not propagated — claim check cleanup is best-effort.
func (c *Consumer) removeClaimCheck(key, orderID string) {
	if err := c.redisClient.Del(context.Background(), key).Err(); err != nil {
		log.Printf("Failed to remove claim check from Redis for order %s (key=%s): %v", orderID, key, err)
	} else {
		log.Printf("Removed claim check from Redis for order %s (key=%s)", orderID, key)
	}
}
