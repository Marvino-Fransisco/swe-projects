package handler

import (
	"net/http"

	"api-gateway/client"
	"api-gateway/model"

	"github.com/gin-gonic/gin"
)

const orderServiceURL = "http://order_devcontainer-app-1:8002"

// OrderHandler handles order-related requests.
type OrderHandler struct {
	client *client.ServiceClient
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		client: client.NewServiceClient(),
	}
}

// CreateOrder handles POST /api/orders
// It forwards the order request to the order service.
// The order service will publish an event to RabbitMQ for the payment service to process.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	resp, err := h.client.Post(orderServiceURL+"/api/orders", req)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Failed to create order",
			Error:   err.Error(),
		})
		return
	}

	if resp.StatusCode != http.StatusCreated {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Failed to create order",
			Error:   string(resp.Body),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", resp.Body)
}
