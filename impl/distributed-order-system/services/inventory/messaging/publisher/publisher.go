package publisher

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewPublisher(amqpURL string) (*Publisher, error) {
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
		"inventories",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare inventories exchange: %w", err)
	}

	if err := ch.ExchangeDeclare(
		"inventories.dlx",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare inventories.dlx exchange: %w", err)
	}

	log.Println("Publisher connected to RabbitMQ successfully")
	return &Publisher{connection: conn, channel: ch}, nil
}

func (p *Publisher) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
}
