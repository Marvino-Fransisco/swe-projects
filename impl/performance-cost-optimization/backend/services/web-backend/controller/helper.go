package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"web-backend/apperror"

	"shared/response"
)

// handleError maps application errors to HTTP responses.
// If the error is an *apperror.Error, it uses the embedded status code.
// Otherwise, it responds with 500 Internal Server Error.
func handleError(c *gin.Context, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		response.Error(c, appErr.Status, appErr.Message)
		return
	}
	response.InternalError(c, err.Error())
}

// ParsePagination extracts page and page_size from query parameters.
// Returns defaults (page=1, pageSize=20) if not provided or invalid.
func ParsePagination(c *gin.Context) (page, pageSize int) {
	page = 1
	pageSize = 20
	if v := c.Query("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := c.Query("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			pageSize = ps
		}
	}
	return page, pageSize
}

// calcTotalPages computes the number of total pages for pagination.
func calcTotalPages(total int64, pageSize int) int {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return totalPages
}

// respondPaginated sends a standard paginated response.
func respondPaginated(c *gin.Context, data any, total int64, page, pageSize int) {
	response.Paginated(c, http.StatusOK, data, response.PaginatedMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: calcTotalPages(total, pageSize),
	})
}
