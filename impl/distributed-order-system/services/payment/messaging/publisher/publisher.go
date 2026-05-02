package publisher

import (
	"fmt"

	sharedRabbitMQ "shared/rabbitmq"
)

// Publisher wraps the shared BasePublisher and declares the payments exchange topology.
type Publisher struct {
	*sharedRabbitMQ.BasePublisher
}

// NewPublisher creates a new Publisher by dialing RabbitMQ and declaring the
// payments exchange with its dead-letter exchange.
func NewPublisher(amqpURL string) (*Publisher, error) {
	base, err := sharedRabbitMQ.NewBasePublisher(amqpURL)
	if err != nil {
		return nil, err
	}

	if err := base.DeclareExchangeWithDLX("payments"); err != nil {
		base.Close()
		return nil, fmt.Errorf("failed to declare payments exchange: %w", err)
	}

	return &Publisher{BasePublisher: base}, nil
}

// Close releases the underlying channel and connection via BasePublisher.
func (p *Publisher) Close() {
	if p.BasePublisher != nil {
		p.BasePublisher.Close()
	}
}
