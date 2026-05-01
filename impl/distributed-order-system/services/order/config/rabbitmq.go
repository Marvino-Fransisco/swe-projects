package config

import "os"

type RabbitMQConfig struct {
	AMQPURL string
}

func DefaultRabbitMQConfig() *RabbitMQConfig {
	url := os.Getenv("AMQP_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	return &RabbitMQConfig{
		AMQPURL: url,
	}
}
