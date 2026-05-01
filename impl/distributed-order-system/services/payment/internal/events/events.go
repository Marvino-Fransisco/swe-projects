package events

import "time"

// PaymentSucceededEvent is published when a payment succeeds.
type PaymentSucceededEvent struct {
	OrderID string `json:"orderId"`
}

// PaymentFailedEvent is published when a payment fails.
type PaymentFailedEvent struct {
	OrderID string `json:"orderId"`
}

// InventoryProduct represents a product within an inventory event.
type InventoryProduct struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// StockReservedEvent is consumed from the inventories exchange.
type StockReservedEvent struct {
	OrderID   string             `json:"orderId"`
	Products  []InventoryProduct `json:"products"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}
