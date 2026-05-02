package publisher

import (
	"fmt"

	sharedRabbitMQ "shared/rabbitmq"
)

type Publisher struct {
	*sharedRabbitMQ.BasePublisher
}

func NewPublisher(amqpURL string) (*Publisher, error) {
	base, err := sharedRabbitMQ.NewBasePublisher(amqpURL)
	if err != nil {
		return nil, err
	}

	if err := base.DeclareExchangeWithDLX("inventories"); err != nil {
		base.Close()
		return nil, fmt.Errorf("failed to declare inventories exchange: %w", err)
	}

	return &Publisher{BasePublisher: base}, nil
}
