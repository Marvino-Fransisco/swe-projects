package bootstrap

import (
	"log"

	sharedRabbitMQ "shared/rabbitmq"
	"inventory-service/internal/app"
	"inventory-service/messaging/consumer"
	"inventory-service/messaging/publisher"

	"github.com/redis/go-redis/v9"
)

// InitPublisher creates the RabbitMQ publisher.
func InitPublisher() *publisher.Publisher {
	cfg := sharedRabbitMQ.DefaultConfig()

	pub, err := publisher.NewPublisher(cfg.AMQPURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ publisher: %v", err)
	}

	return pub
}

// InitRabbitMQ creates the consumer and starts listening for events.
// It uses the application layer to handle commands.
func InitRabbitMQ(application *app.Application, redisClient *redis.Client) *consumer.Consumer {
	cfg := sharedRabbitMQ.DefaultConfig()

	con, err := consumer.NewConsumer(cfg.AMQPURL, application, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}

	if err := con.StartConsumingOrderEvents(); err != nil {
		log.Fatalf("Failed to start consuming order events: %v", err)
	}

	if err := con.StartConsumingPaymentEvents(); err != nil {
		log.Fatalf("Failed to start consuming payment events: %v", err)
	}

	return con
}
