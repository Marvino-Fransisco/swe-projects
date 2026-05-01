package routes

import (
	"inventory-service/internal/adapters/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures the API routes for the inventory service.
func SetupRoutes(router *gin.Engine, handler *http.InventoryHandler) {
	api := router.Group("/api")
	{
		api.GET("/inventories", handler.GetInventories)
	}
}
