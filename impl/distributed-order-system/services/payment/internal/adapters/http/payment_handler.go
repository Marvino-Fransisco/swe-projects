package http

import (
	"log"
	"net/http"

	"payment-service/internal/app"
	"payment-service/internal/app/command"
	"payment-service/internal/domain/payment"

	"github.com/gin-gonic/gin"
)

// PaymentHandler is a thin HTTP adapter. It parses requests, delegates to the
// application layer, and serializes responses. No business logic lives here.
type PaymentHandler struct {
	app *app.Application
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(app *app.Application) *PaymentHandler {
	return &PaymentHandler{app: app}
}

// ProcessPayment handles POST /api/payments/:payment_id/process.
func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	paymentID := c.Param("payment_id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, APIError{
			Status:  http.StatusBadRequest,
			Message: "Payment ID is required",
			Error:   "missing payment_id parameter",
		})
		return
	}

	var req ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIError{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	cmd := command.ProcessPayment{
		PaymentID: paymentID,
		Amount:    req.Amount,
	}

	result, err := h.app.Commands.ProcessPayment.Handle(c.Request.Context(), cmd)
	if err != nil {
		log.Printf("Failed to process payment %s: %v", paymentID, err)
		c.JSON(http.StatusInternalServerError, APIError{
			Status:  http.StatusInternalServerError,
			Message: "Failed to process payment",
			Error:   err.Error(),
		})
		return
	}

	if result.Status == payment.StatusSucceeded {
		c.JSON(http.StatusOK, toPaymentResponse(result))
	} else {
		c.JSON(http.StatusBadRequest, APIError{
			Status:  http.StatusBadRequest,
			Message: "Insufficient payment amount",
			Error:   "payment amount is less than total price",
		})
	}
}

// toPaymentResponse converts a command result to an HTTP PaymentResponse.
func toPaymentResponse(r command.ProcessPaymentResult) PaymentResponse {
	return PaymentResponse{
		PaymentID:  r.PaymentID,
		OrderID:    r.OrderID,
		TotalPrice: r.TotalPrice,
		Status:     string(r.Status),
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
