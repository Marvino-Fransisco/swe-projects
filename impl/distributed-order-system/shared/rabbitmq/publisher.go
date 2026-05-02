package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// BasePublisher holds the RabbitMQ connection and channel.
// Services should embed this struct and add their business-specific publish methods.
//
// Example:
//
//	type OrderPublisher struct {
//	    *rabbitmq.BasePublisher
//	    redisClient *redis.Client
//	}
//
//	func (p *OrderPublisher) PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
//	    // business logic using p.Channel to publish
//	}
type BasePublisher struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

// NewBasePublisher dials RabbitMQ and opens a channel.
// Call DeclareExchange or DeclareExchangeWithDLX to set up the required topology,
// then use the Channel field to publish messages.
func NewBasePublisher(amqpURL string) (*BasePublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	log.Println("Publisher connected to RabbitMQ successfully")
	return &BasePublisher{Connection: conn, Channel: ch}, nil
}

// DeclareExchange declares a durable topic exchange.
func (p *BasePublisher) DeclareExchange(name string) error {
	if err := p.Channel.ExchangeDeclare(
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
func (p *BasePublisher) DeclareExchangeWithDLX(name string) error {
	if err := p.DeclareExchange(name); err != nil {
		return err
	}

	dlxName := name + ".dlx"
	if err := p.Channel.ExchangeDeclare(
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

// Close releases the channel and connection.
func (p *BasePublisher) Close() {
	if p.Channel != nil {
		p.Channel.Close()
	}
	if p.Connection != nil {
		p.Connection.Close()
	}
}
