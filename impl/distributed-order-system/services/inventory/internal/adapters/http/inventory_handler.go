package http

import (
	"math"
	"net/http"
	"strconv"

	"inventory-service/internal/app"
	"inventory-service/internal/app/query"

	"github.com/gin-gonic/gin"
)

// InventoryHandler is a thin HTTP adapter. It parses requests, delegates to the
// application layer, and serializes responses. No business logic lives here.
type InventoryHandler struct {
	app *app.Application
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler(app *app.Application) *InventoryHandler {
	return &InventoryHandler{app: app}
}

// GetInventories handles GET /api/inventories.
// It returns a paginated list of inventory items.
func (h *InventoryHandler) GetInventories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	result, err := h.app.Queries.ListInventories.Handle(c.Request.Context(), query.ListInventories{
		Page:  page,
		Limit: limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIError{
			Status:  http.StatusInternalServerError,
			Message: "Failed to fetch inventories",
			Error:   err.Error(),
		})
		return
	}

	responses := make([]InventoryResponse, 0, len(result.Data))
	for _, v := range result.Data {
		responses = append(responses, toInventoryResponse(v))
	}

	totalPages := int(math.Ceil(float64(result.Total) / float64(limit)))

	c.JSON(http.StatusOK, PaginationResponse{
		Data:       responses,
		Total:      result.Total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// toInventoryResponse converts a query InventoryView to an HTTP InventoryResponse.
func toInventoryResponse(v query.InventoryView) InventoryResponse {
	return InventoryResponse{
		ID:          v.ID,
		ProductID:   v.ProductID,
		ProductName: v.ProductName,
		Stock:       v.Stock,
		Price:       v.Price,
		Status:      v.Status,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}
