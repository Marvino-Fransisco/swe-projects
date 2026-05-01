package routes

import (
	"payment-service/internal/adapters/http"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, handler *http.PaymentHandler) {
	api := router.Group("/api")
	{
		api.POST("/payments/:payment_id/process", handler.ProcessPayment)
	}
}
