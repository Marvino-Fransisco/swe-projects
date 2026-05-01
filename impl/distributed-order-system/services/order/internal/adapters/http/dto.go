package http

import "time"

// CreateOrderRequest is the HTTP request DTO for creating an order.
type CreateOrderRequest struct {
	Products []CreateOrderProduct `json:"products" binding:"required,dive"`
}

// CreateOrderProduct is a product item within a CreateOrderRequest.
type CreateOrderProduct struct {
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// OrderResponse is the HTTP response DTO for an order.
type OrderResponse struct {
	ID            string            `json:"id"`
	Products      []ProductResponse `json:"products"`
	Status        string            `json:"status"`
	FailureReason string            `json:"failureReason"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

// ProductResponse is the HTTP response DTO for an order product.
type ProductResponse struct {
	ID        string    `json:"id"`
	ProductID string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// APIError is the HTTP response DTO for errors.
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}
