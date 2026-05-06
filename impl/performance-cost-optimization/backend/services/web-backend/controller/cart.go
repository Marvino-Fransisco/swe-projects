package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shared/middleware"
	"shared/response"

	"web-backend/usecases/cart"
)

// CartController defines the interface for cart HTTP handlers.
type CartController interface {
	GetCart(c *gin.Context)
	AddItem(c *gin.Context)
	RemoveItem(c *gin.Context)
	UpdateItemQuantity(c *gin.Context)
}

type cartController struct {
	usecase cart.CartUseCase
}

// NewCartController creates a new CartController.
func NewCartController(usecase cart.CartUseCase) CartController {
	return &cartController{usecase: usecase}
}

// GetCart handles GET /cart.
func (ctrl *cartController) GetCart(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	page, pageSize := ParsePagination(c)

	resp, err := ctrl.usecase.GetCart(c.Request.Context(), cart.GetCartRequest{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "cart retrieved", resp.Items)
}

// AddItem handles POST /cart/items.
func (ctrl *cartController) AddItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req cart.AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.UserID = userID

	resp, err := ctrl.usecase.AddItem(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item added to cart", resp.Item)
}

// RemoveItem handles DELETE /cart/items/:productId.
func (ctrl *cartController) RemoveItem(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		response.BadRequest(c, "product id is required")
		return
	}

	resp, err := ctrl.usecase.RemoveItem(c.Request.Context(), cart.RemoveCartItemRequest{
		UserID:    userID,
		ProductID: productID,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "item removed from cart", resp.Item)
}

// UpdateItemQuantity handles PUT /cart/items/:productId.
func (ctrl *cartController) UpdateItemQuantity(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		response.BadRequest(c, "product id is required")
		return
	}

	var req cart.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.UserID = userID
	req.ProductID = productID

	resp, err := ctrl.usecase.UpdateItemQuantity(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "cart item updated", resp.Item)
}
