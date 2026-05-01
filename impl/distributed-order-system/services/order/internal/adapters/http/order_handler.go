package http

import (
	"log"
	"net/http"

	"order-service/internal/app"
	"order-service/internal/app/command"
	"order-service/internal/app/query"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderHandler is a thin HTTP adapter. It parses requests, delegates to the
// application layer, and serializes responses. No business logic lives here.
type OrderHandler struct {
	app *app.Application
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(app *app.Application) *OrderHandler {
	return &OrderHandler{app: app}
}

// CreateOrder handles POST /api/orders.
// It delegates to the CreateOrder command, then queries the created order.
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIError{
			Status:  http.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	orderID := uuid.New().String()

	products := make([]command.CreateOrderProduct, 0, len(req.Products))
	for _, p := range req.Products {
		products = append(products, command.CreateOrderProduct{
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
		})
	}

	cmd := command.CreateOrder{
		OrderID:  orderID,
		Products: products,
	}

	if err := h.app.Commands.CreateOrder.Handle(c.Request.Context(), cmd); err != nil {
		log.Printf("Failed to create order: %v", err)
		c.JSON(http.StatusInternalServerError, APIError{
			Status:  http.StatusInternalServerError,
			Message: "Failed to create order",
			Error:   err.Error(),
		})
		return
	}

	// CQRS: after the write succeeds, use the read side to fetch the created order.
	view, err := h.app.Queries.GetOrder.Handle(c.Request.Context(), query.GetOrder{ID: orderID})
	if err != nil {
		log.Printf("Order %s created but failed to fetch: %v", orderID, err)
		c.Header("Content-Location", "/api/orders/"+orderID)
		c.Status(http.StatusCreated)
		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(view))
}

// GetOrder handles GET /api/orders/:id.
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	view, err := h.app.Queries.GetOrder.Handle(c.Request.Context(), query.GetOrder{ID: id})
	if err != nil {
		c.JSON(http.StatusNotFound, APIError{
			Status:  http.StatusNotFound,
			Message: "Order not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, toOrderResponse(view))
}

// ListOrders handles GET /api/orders.
func (h *OrderHandler) ListOrders(c *gin.Context) {
	views, err := h.app.Queries.ListOrders.Handle(c.Request.Context(), query.ListOrders{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIError{
			Status:  http.StatusInternalServerError,
			Message: "Failed to list orders",
			Error:   err.Error(),
		})
		return
	}

	responses := make([]OrderResponse, 0, len(views))
	for _, v := range views {
		responses = append(responses, toOrderResponse(v))
	}

	c.JSON(http.StatusOK, responses)
}

// toOrderResponse converts a query OrderView to an HTTP OrderResponse.
func toOrderResponse(v query.OrderView) OrderResponse {
	products := make([]ProductResponse, 0, len(v.Products))
	for _, p := range v.Products {
		products = append(products, ProductResponse{
			ID:        p.ID,
			ProductID: p.ProductID,
			Quantity:  p.Quantity,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}
	return OrderResponse{
		ID:            v.ID,
		Products:      products,
		Status:        v.Status,
		FailureReason: v.FailureReason,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
	}
}
