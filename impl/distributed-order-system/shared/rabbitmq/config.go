package rabbitmq

import "os"

// RabbitMQConfig holds the RabbitMQ connection configuration.
type RabbitMQConfig struct {
	AMQPURL string
}

// DefaultConfig returns the default RabbitMQ configuration.
// It reads AMQP_URL from the environment, falling back to the default guest connection.
func DefaultConfig() *RabbitMQConfig {
	url := os.Getenv("AMQP_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	return &RabbitMQConfig{
		AMQPURL: url,
	}
}
