package dtos

type ErrorCode string

const (
	ErrorCodeServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrorCodeResourceNotFound    ErrorCode = "RESOURCE_NOT_FOUND"
	ErrorCodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeTimeout             ErrorCode = "TIMEOUT"
	ErrorTCPConnectionTimeout    ErrorCode = "TCP_CONNECTION_TIMEOUT"
	ErrorCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden           ErrorCode = "FORBIDDEN"
	ErrorCodeMaxRetriesExceeded  ErrorCode = "MAX_RETRIES_EXCEEDED"
	ErrorCodeCapacityExceeded    ErrorCode = "CAPACITY_EXCEEDED"
)

type ErrorResponse struct {
	Success   bool      `json:"success"`
	ErrorCode ErrorCode `json:"error_code"`
	Message   string    `json:"message"`
	RequestID string    `json:"request_id,omitempty"`
}
