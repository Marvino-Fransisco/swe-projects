package events

import "time"

// StockReservedEvent is published when all products in an order have sufficient stock.
type StockReservedEvent struct {
	OrderID   string             `json:"orderId"`
	Products  []InventoryProduct `json:"products"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// StockRejectedEvent is published when one or more products have insufficient stock or are not found.
type StockRejectedEvent struct {
	OrderID   string             `json:"orderId"`
	Products  []InventoryProduct `json:"products"`
	Status    string             `json:"status"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

// InventoryProduct represents a product within an inventory event.
type InventoryProduct struct {
	ProductID string  `json:"productId"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
