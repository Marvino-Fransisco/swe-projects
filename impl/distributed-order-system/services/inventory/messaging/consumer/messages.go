package consumer

// ClaimCheckMessage is the lightweight message received from RabbitMQ.
// The full payload is stored in Supabase Storage and must be fetched using the ClaimCheckKey.
type ClaimCheckMessage struct {
	OrderID       string `json:"orderId"`
	ClaimCheckKey string `json:"claimCheckKey"`
}

// OrderCreatedPayload represents the full event payload fetched from Supabase Storage.
type OrderCreatedPayload struct {
	ID        string                    `json:"id"`
	Products  []OrderProductPayload     `json:"products"`
	Status    string                    `json:"status"`
	CreatedAt string                    `json:"createdAt"`
	UpdatedAt string                    `json:"updatedAt"`
}

// OrderProductPayload represents a product within the full OrderCreatedPayload.
type OrderProductPayload struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// PaymentMessage is deserialized from incoming payment events.
type PaymentMessage struct {
	OrderID string `json:"orderId"`
}
