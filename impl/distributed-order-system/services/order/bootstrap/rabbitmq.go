package bootstrap

import (
	"log"

	"order-service/internal/app"
	"order-service/messaging/consumer"
	"order-service/messaging/publisher"

	sharedRabbitMQ "shared/rabbitmq"

	"github.com/redis/go-redis/v9"
)

// InitPublisher creates the RabbitMQ publisher with the shared Redis client.
func InitPublisher(redisClient *redis.Client) *publisher.Publisher {
	cfg := sharedRabbitMQ.DefaultConfig()

	pub, err := publisher.NewPublisher(cfg.AMQPURL, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ publisher: %v", err)
	}

	return pub
}

// InitRabbitMQ creates the consumer and starts listening for events.
// It uses the application layer to handle status update commands.
func InitRabbitMQ(application *app.Application) *consumer.Consumer {
	cfg := sharedRabbitMQ.DefaultConfig()

	con, err := consumer.NewConsumer(cfg.AMQPURL, application)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}

	if err := con.StartConsumingPaymentEvents(); err != nil {
		log.Fatalf("Failed to start consuming payment events: %v", err)
	}

	if err := con.StartConsumingInventoryEvents(); err != nil {
		log.Fatalf("Failed to start consuming inventory events: %v", err)
	}

	return con
}
