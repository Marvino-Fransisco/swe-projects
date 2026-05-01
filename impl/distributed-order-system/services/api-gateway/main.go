package main

import (
	"log"

	"api-gateway/handler"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	inventoryHandler := handler.NewInventoryHandler()
	orderHandler := handler.NewOrderHandler()
	paymentHandler := handler.NewPaymentHandler()
	webhookHandler := handler.NewWebhookHandler()

	api := router.Group("/api")
	{
		api.GET("/inventories", inventoryHandler.GetInventories)
		api.POST("/orders", orderHandler.CreateOrder)
		api.POST("/payments/:payment_id/process", paymentHandler.ProcessPayment)
		api.POST("/webhooks", webhookHandler.ReceiveWebhook)
	}

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
