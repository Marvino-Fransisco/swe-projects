package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"shared/response"

	usecaseProduct "mobile-backend/usecases/product"
)

// ProductController defines the interface for product HTTP handlers.
// It handles both product catalog operations and product view tracking.
type ProductController interface {
	GetByID(c *gin.Context)
	List(c *gin.Context)
	Search(c *gin.Context)
	GetCategories(c *gin.Context)
	TrackView(c *gin.Context)
}

type productController struct {
	usecase usecaseProduct.ProductUseCase
}

// NewProductController creates a new ProductController.
func NewProductController(usecase usecaseProduct.ProductUseCase) ProductController {
	return &productController{usecase: usecase}
}

// GetByID handles GET /products/:id.
func (ctrl *productController) GetByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "product id is required")
		return
	}

	p, err := ctrl.usecase.GetByID(c.Request.Context(), usecaseProduct.GetProductRequest{ID: id})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "product retrieved", p)
}

// List handles GET /products.
func (ctrl *productController) List(c *gin.Context) {
	page, pageSize := ParsePagination(c)

	req := usecaseProduct.ListProductsRequest{
		Page:     page,
		PageSize: pageSize,
	}

	// Apply optional filters from query parameters.
	if v := c.Query("category_id"); v != "" {
		req.CategoryID = v
	}
	if v := c.Query("search"); v != "" {
		req.Search = v
	}
	if v := c.Query("sort_by"); v != "" {
		req.SortBy = v
	}
	if v := c.Query("sort_order"); v != "" {
		req.SortOrder = v
	}
	if v := c.Query("min_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.MinPrice = f
		}
	}
	if v := c.Query("max_price"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			req.MaxPrice = f
		}
	}

	resp, err := ctrl.usecase.List(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	respondPaginated(c, resp.Products, resp.Total, page, pageSize)
}

// Search handles GET /products/search.
func (ctrl *productController) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		response.BadRequest(c, "search query 'q' is required")
		return
	}

	page, pageSize := ParsePagination(c)

	resp, err := ctrl.usecase.Search(c.Request.Context(), usecaseProduct.SearchProductsRequest{
		Query:    query,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	respondPaginated(c, resp.Products, resp.Total, page, pageSize)
}

// GetCategories handles GET /products/categories.
func (ctrl *productController) GetCategories(c *gin.Context) {
	categories, err := ctrl.usecase.GetCategories(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "categories retrieved", categories)
}

// TrackView handles POST /products/:id/view.
func (ctrl *productController) TrackView(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "product id is required")
		return
	}

	resp, err := ctrl.usecase.TrackView(c.Request.Context(), usecaseProduct.TrackViewRequest{
		ProductID: id,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "view tracked", resp)
}
