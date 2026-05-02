package publisher

import (
	"fmt"
	"log"

	sharedRabbitMQ "shared/rabbitmq"

	"github.com/redis/go-redis/v9"
)

type Publisher struct {
	*sharedRabbitMQ.BasePublisher
	redisClient *redis.Client
}

func NewPublisher(amqpURL string, redisClient *redis.Client) (*Publisher, error) {
	base, err := sharedRabbitMQ.NewBasePublisher(amqpURL)
	if err != nil {
		return nil, err
	}

	if err := base.DeclareExchangeWithDLX("orders"); err != nil {
		base.Close()
		return nil, fmt.Errorf("failed to declare orders exchange: %w", err)
	}

	log.Println("Publisher connected to RabbitMQ successfully")
	return &Publisher{
		BasePublisher: base,
		redisClient:   redisClient,
	}, nil
}
