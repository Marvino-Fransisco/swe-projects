package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"api-gateway/client"
	"api-gateway/model"

	"github.com/gin-gonic/gin"
)

const inventoryServiceURL = "http://inventory_devcontainer-app-1:8001"

// InventoryHandler handles inventory-related requests.
type InventoryHandler struct {
	client *client.ServiceClient
}

// NewInventoryHandler creates a new InventoryHandler.
func NewInventoryHandler() *InventoryHandler {
	return &InventoryHandler{
		client: client.NewServiceClient(),
	}
}

// GetInventories handles GET /api/inventories
// It fetches paginated inventory data from the inventory service.
func (h *InventoryHandler) GetInventories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	url := fmt.Sprintf("%s/api/inventories?page=%d&limit=%d", inventoryServiceURL, page, limit)

	resp, err := h.client.Get(url)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Failed to fetch inventories",
			Error:   err.Error(),
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Failed to fetch inventories",
			Error:   string(resp.Body),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", resp.Body)
}
