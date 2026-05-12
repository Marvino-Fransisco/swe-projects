package routes

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	order_http "order-service/internal/adapters/http"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

func SetupRoutes(router *gin.Engine, handler *order_http.OrderHandler) {
	router.GET("/health", healthHandler)

	api := router.Group("/api")
	{
		api.POST("/orders", handler.CreateOrder)
		api.GET("/orders", handler.ListOrders)
		api.GET("/orders/:id", handler.GetOrder)
	}
}

func healthHandler(c *gin.Context) {
	hostname, _ := os.Hostname()

	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"service":    "order-service",
		"hostname":   hostname,
		"go_version": runtime.Version(),
		"uptime":     time.Since(startTime).String(),
		"version":    os.Getenv("VERSION"),
		"checked_at": time.Now().UTC().Format(time.RFC3339),
		"gcp": gin.H{
			"project":      os.Getenv("GCP_PROJECT"),
			"region":       os.Getenv("GCP_REGION"),
			"service_name": os.Getenv("K_SERVICE"),
		},
		"metadata": gin.H{
			"pid":        os.Getpid(),
			"goroutines": runtime.NumGoroutine(),
			"cpu_count":  runtime.NumCPU(),
		},
		"message": fmt.Sprintf("healthy — uptime %s on %s", time.Since(startTime).Round(time.Second), hostname),
	})
}
