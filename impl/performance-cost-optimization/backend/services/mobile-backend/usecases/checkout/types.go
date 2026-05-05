package checkout

import "time"

type PlaceOrderRequest struct {
	UserID string
}

type GetOrderRequest struct {
	UserID  string
	OrderID string
}

type GetOrderHistoryRequest struct {
	UserID   string
	Page     int
	PageSize int
}

// OrderResponse is a slim representation of an order for mobile clients.
type OrderResponse struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	FailureReason string    `json:"failure_reason,omitempty"`
}

// OrderItemResponse is a slim representation of an order line item for mobile clients.
type OrderItemResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderWithDetailsResponse struct {
	Order  *OrderResponse      `json:"order"`
	Items  []OrderItemResponse `json:"items"`
}

type OrderHistoryResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int64           `json:"total"`
}
