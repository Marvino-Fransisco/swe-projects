package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard success response wrapper.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// APIError is the standard error response.
type APIError struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// PaginatedMeta holds pagination metadata.
type PaginatedMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// PaginatedResponse wraps a paginated list response.
type PaginatedResponse struct {
	Success bool         `json:"success"`
	Data    any          `json:"data"`
	Meta    PaginatedMeta `json:"meta"`
}

// Success sends a success JSON response.
func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error JSON response.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, APIError{
		Success: false,
		Message: message,
	})
}

// ErrorWithCode sends an error JSON response with a specific error code.
func ErrorWithCode(c *gin.Context, status int, code, message string) {
	c.JSON(status, APIError{
		Success: false,
		Message: message,
		Code:    code,
	})
}

// Paginated sends a paginated list JSON response.
func Paginated(c *gin.Context, status int, data any, meta PaginatedMeta) {
	c.JSON(status, PaginatedResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Unauthorized sends a 401 unauthorized response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 forbidden response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 not found response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// BadRequest sends a 400 bad request response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// InternalError sends a 500 internal server error response.
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
