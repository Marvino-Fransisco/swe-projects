package publisher

import (
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	connection  *amqp.Connection
	channel     *amqp.Channel
	redisClient *redis.Client
}

func NewPublisher(amqpURL string, redisClient *redis.Client) (*Publisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		"orders",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare orders exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(
		"orders.dlx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare orders.retry exchange: %w", err)
	}

	log.Println("Publisher connected to RabbitMQ successfully")
	return &Publisher{
		connection:  conn,
		channel:     ch,
		redisClient: redisClient,
	}, nil
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
}
