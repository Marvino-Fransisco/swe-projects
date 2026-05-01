package routes

import (
	"order-service/internal/adapters/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handler *http.OrderHandler) {
	api := router.Group("/api")
	{
		api.POST("/orders", handler.CreateOrder)
		api.GET("/orders", handler.ListOrders)
		api.GET("/orders/:id", handler.GetOrder)
	}
}
