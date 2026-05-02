package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DefaultMaxRetries is the default maximum number of retry attempts
// before a message is sent to the dead-letter queue.
const DefaultMaxRetries int32 = 5

// BaseConsumer holds the RabbitMQ connection and channel.
// Services should embed this struct and add their business-specific consume methods.
//
// Example:
//
//	type OrderConsumer struct {
//	    *rabbitmq.BaseConsumer
//	    app         *app.Application
//	    redisClient *redis.Client
//	}
//
//	func (c *OrderConsumer) StartConsumingInventoryEvents() error {
//	    msgs, err := c.Channel.Consume(...)
//	    // business logic for handling messages
//	}
type BaseConsumer struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewBaseConsumer dials RabbitMQ and opens a channel.
// Call DeclareExchangeWithDLX, DeclareQueue, and BindQueue to set up
// the required topology, then use the Channel field to consume messages.
func NewBaseConsumer(amqpURL string) (*BaseConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	log.Println("Consumer connected to RabbitMQ successfully")
	return &BaseConsumer{Connection: conn, Channel: ch}, nil
}

// DeclareExchange declares a durable topic exchange.
func (c *BaseConsumer) DeclareExchange(name string) error {
	if err := c.Channel.ExchangeDeclare(
		name,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare %s exchange: %w", name, err)
	}
	return nil
}

// DeclareExchangeWithDLX declares a durable topic exchange along with its
// dead-letter exchange (<name>.dlx). This is the standard pattern used
// across all services for retry/DLQ support.
func (c *BaseConsumer) DeclareExchangeWithDLX(name string) error {
	if err := c.DeclareExchange(name); err != nil {
		return err
	}

	dlxName := name + ".dlx"
	if err := c.Channel.ExchangeDeclare(
		dlxName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare %s exchange: %w", dlxName, err)
	}

	return nil
}

// DeclareQueue declares a queue with the given name and optional arguments.
// Common arguments include x-dead-letter-exchange, x-dead-letter-routing-key,
// and x-message-ttl for retry/DLQ topologies.
func (c *BaseConsumer) DeclareQueue(name string, args amqp.Table) (amqp.Queue, error) {
	q, err := c.Channel.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare %s queue: %w", name, err)
	}
	return q, nil
}

// BindQueue binds a queue to an exchange with the given routing key.
func (c *BaseConsumer) BindQueue(queueName, routingKey, exchange string) error {
	if err := c.Channel.QueueBind(
		queueName,
		routingKey,
		exchange,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to bind %s to %s exchange: %w", queueName, exchange, err)
	}
	return nil
}

// PublishToRetry publishes a message to a retry exchange with the retry count
// stored in the x-retry-count header.
func (c *BaseConsumer) PublishToRetry(exchange, routingKey string, body []byte, retryCount int32) error {
	return c.Channel.Publish(
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

// PublishToDLQ publishes a message to a dead-letter exchange.
func (c *BaseConsumer) PublishToDLQ(exchange, routingKey string, body []byte) error {
	return c.Channel.Publish(
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

// GetRetryCount extracts the retry count from the x-retry-count header.
// Returns 0 if the header is missing or cannot be parsed.
func GetRetryCount(headers amqp.Table) int32 {
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

// Close releases the channel and connection.
func (c *BaseConsumer) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.Connection != nil {
		c.Connection.Close()
	}
}
