package http

import "time"

// APIError is the HTTP response DTO for errors.
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// InventoryResponse is the HTTP response DTO for an inventory item.
type InventoryResponse struct {
	ID          uint      `json:"id"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	Stock       int       `json:"quantity"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PaginationResponse is the paginated HTTP response for inventory list.
type PaginationResponse struct {
	Data       []InventoryResponse `json:"data"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"totalPages"`
}
