package apperror

// Error represents an application-level error with an HTTP status code.
// Use case layers return these so the controller knows the appropriate HTTP status.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }

// NewBadRequest creates a 400 Bad Request error.
func NewBadRequest(msg string) *Error { return &Error{Status: 400, Message: msg} }

// NewNotFound creates a 404 Not Found error.
func NewNotFound(msg string) *Error { return &Error{Status: 404, Message: msg} }

// NewUnauthorized creates a 401 Unauthorized error.
func NewUnauthorized(msg string) *Error { return &Error{Status: 401, Message: msg} }
