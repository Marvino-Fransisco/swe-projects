package bootstrap

import (
	"log"

	"payment-service/config"
	"payment-service/internal/app"
	"payment-service/messaging/consumer"
	"payment-service/messaging/publisher"
)

// InitPublisher creates the RabbitMQ publisher.
func InitPublisher() *publisher.Publisher {
	cfg := config.DefaultRabbitMQConfig()

	pub, err := publisher.NewPublisher(cfg.AMQPURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ publisher: %v", err)
	}

	return pub
}

// InitRabbitMQ creates the consumer and starts listening for events.
// It uses the application layer to handle payment creation commands.
func InitRabbitMQ(application *app.Application) *consumer.Consumer {
	cfg := config.DefaultRabbitMQConfig()

	con, err := consumer.NewConsumer(cfg.AMQPURL, application)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}

	if err := con.StartConsumingInventoryEvents(); err != nil {
		log.Fatalf("Failed to start consuming inventory events: %v", err)
	}

	return con
}
