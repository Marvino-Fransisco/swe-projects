package handler

import (
	"fmt"
	"net/http"

	"api-gateway/client"
	"api-gateway/model"

	"github.com/gin-gonic/gin"
)

const paymentServiceURL = "http://payment_devcontainer-app-1:8003"

type PaymentHandler struct {
	client *client.ServiceClient
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{
		client: client.NewServiceClient(),
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")

	var req model.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	url := fmt.Sprintf("%s/api/payments/%s/process", paymentServiceURL, paymentID)

	resp, err := h.client.Post(url, req)
	if err != nil {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Failed to process payment",
			Error:   err.Error(),
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusNotFound, model.APIError{
			Status:  http.StatusNotFound,
			Message: "Failed to process payment",
			Error:   string(resp.Body),
		})
		return
	}

	c.Data(http.StatusOK, "application/json", resp.Body)
}
