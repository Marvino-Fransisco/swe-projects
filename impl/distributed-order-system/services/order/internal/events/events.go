package events

import "time"

// OrderCreatedEvent is published when a new order is created.
type OrderCreatedEvent struct {
	ID        string              `json:"id"`
	Products  []OrderProductEvent `json:"products"`
	Status    string              `json:"status"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

// OrderProductEvent represents a product within an OrderCreatedEvent.
type OrderProductEvent struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}
