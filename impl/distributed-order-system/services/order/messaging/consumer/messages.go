package consumer

// PaymentMessage is deserialized from incoming payment events.
type PaymentMessage struct {
	OrderID string `json:"orderId"`
}

// InventoryMessage is deserialized from incoming inventory events.
type InventoryMessage struct {
	OrderID string `json:"orderId"`
}
