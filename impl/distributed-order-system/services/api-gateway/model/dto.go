package model

// CreateOrderRequest is the request body for creating an order.
type CreateOrderRequest struct {
	Products []Product
}

type Product struct {
	ProductId string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// ProcessPaymentRequest is the request body for processing a payment.
type ProcessPaymentRequest struct {
	Amount float64 `json:"amount"`
}

// APIError is the standard error response format.
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}
