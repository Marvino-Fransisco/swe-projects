package http

import "time"

// ProcessPaymentRequest is the HTTP request DTO for processing a payment.
type ProcessPaymentRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// PaymentResponse is the HTTP response DTO for a payment.
type PaymentResponse struct {
	PaymentID  string    `json:"paymentId"`
	OrderID    string    `json:"orderId"`
	TotalPrice float64   `json:"totalPrice"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// APIError is the HTTP response DTO for errors.
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}
