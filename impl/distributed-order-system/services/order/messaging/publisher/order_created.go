package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"order-service/internal/events"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	claimCheckTTL = time.Hour
)

// claimCheckMessage is the lightweight message sent to RabbitMQ after
// the full payload has been stored in Redis.
type claimCheckMessage struct {
	OrderID       string `json:"orderId"`
	ClaimCheckKey string `json:"claimCheckKey"`
}

// PublishOrderCreated publishes an OrderCreated event using the Claim Check pattern.
// 1. The full event payload is stored in Redis with a 1-hour TTL.
// 2. Only the orderId and claimCheckKey are sent to RabbitMQ.
// If the Redis store fails, no message is sent to RabbitMQ and the error is returned.
func (p *Publisher) PublishOrderCreated(ctx context.Context, event events.OrderCreatedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	claimCheckKey := fmt.Sprintf("claim-check:orders:%s", event.ID)

	// Step 1: Store full payload in Redis with 1-hour TTL.
	err = p.redisClient.Set(ctx, claimCheckKey, body, claimCheckTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to store event payload in Redis: %w", err)
	}

	// Step 2: Send lightweight claim check message to RabbitMQ.
	claimCheck := claimCheckMessage{
		OrderID:       event.ID,
		ClaimCheckKey: claimCheckKey,
	}

	claimCheckBody, err := json.Marshal(claimCheck)
	if err != nil {
		return fmt.Errorf("failed to marshal claim check message: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"orders",
		"orders.created",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         claimCheckBody,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish claim check message to RabbitMQ: %w", err)
	}

	log.Printf("Published OrderCreated claim check for order %s (key=%s)", event.ID, claimCheckKey)
	return nil
}
