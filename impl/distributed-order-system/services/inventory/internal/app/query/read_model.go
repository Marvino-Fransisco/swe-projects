package query

import (
	"context"
	"time"
)

// InventoryView is the read-side DTO returned by queries.
// It is optimized for the presentation layer and contains no domain logic.
type InventoryView struct {
	ID          uint      `json:"id"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	Stock       int       `json:"quantity"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PaginationResult holds the paginated result for list queries.
type PaginationResult struct {
	Data       []InventoryView `json:"data"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"totalPages"`
}

// InventoryReadModel is the port for querying inventory data.
// The read side (queries) uses this interface.
type InventoryReadModel interface {
	ListInventories(ctx context.Context, page, limit int) (PaginationResult, error)
}
