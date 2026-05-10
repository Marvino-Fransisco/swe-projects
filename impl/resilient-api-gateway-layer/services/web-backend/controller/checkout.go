package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shared/middleware"
	"shared/response"

	"web-backend/usecases/checkout"
)

// CheckoutController defines the interface for checkout HTTP handlers.
type CheckoutController interface {
	PlaceOrder(c *gin.Context)
	GetOrder(c *gin.Context)
	GetOrderHistory(c *gin.Context)
}

type checkoutController struct {
	usecase checkout.CheckoutUseCase
}

// NewCheckoutController creates a new CheckoutController.
func NewCheckoutController(usecase checkout.CheckoutUseCase) CheckoutController {
	return &checkoutController{usecase: usecase}
}

// PlaceOrder handles POST /checkout/orders.
func (ctrl *checkoutController) PlaceOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	o, err := ctrl.usecase.PlaceOrder(c.Request.Context(), checkout.PlaceOrderRequest{
		UserID: userID,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "order placed", o)
}

// GetOrder handles GET /checkout/orders/:id.
func (ctrl *checkoutController) GetOrder(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	orderID := c.Param("id")
	if orderID == "" {
		response.BadRequest(c, "order id is required")
		return
	}

	result, err := ctrl.usecase.GetOrder(c.Request.Context(), checkout.GetOrderRequest{
		UserID:  userID,
		OrderID: orderID,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "order retrieved", result)
}

// GetOrderHistory handles GET /checkout/orders.
func (ctrl *checkoutController) GetOrderHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	page, pageSize := ParsePagination(c)

	resp, err := ctrl.usecase.GetOrderHistory(c.Request.Context(), checkout.GetOrderHistoryRequest{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	respondPaginated(c, resp.Orders, resp.Total, page, pageSize)
}
