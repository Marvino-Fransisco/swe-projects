package middleware

import (
	"github.com/gin-gonic/gin"
)

// Context keys for storing authenticated user data in the Gin context.
const (
	UserIDKey string = "user_id"
	RoleKey   string = "role"
)

// AuthMiddleware defines the interface for authentication middleware.
// Each BFF service (website, mobile) provides its own implementation
// to handle the different token extraction strategies:
//   - Website: extracts token from HTTP-only cookies
//   - Mobile: extracts token from Authorization header
type AuthMiddleware interface {
	// RequireAuth returns a Gin middleware that validates the access token
	// and aborts with 401 if the token is missing or invalid.
	RequireAuth() gin.HandlerFunc

	// OptionalAuth returns a Gin middleware that attempts to validate the
	// access token but does not abort if it is missing or invalid.
	// This allows endpoints that work for both authenticated and
	// unauthenticated users (e.g., product listing).
	OptionalAuth() gin.HandlerFunc
}

// GetUserID extracts the authenticated user ID from the Gin context.
// Returns an empty string if no user is authenticated.
func GetUserID(c *gin.Context) string {
	return c.GetString(UserIDKey)
}

// GetRole extracts the authenticated user's role from the Gin context.
// Returns an empty string if no user is authenticated.
func GetRole(c *gin.Context) string {
	return c.GetString(RoleKey)
}

// SetAuthData stores the authenticated user's ID and role in the Gin context.
// This should be called by the BFF-specific AuthMiddleware implementation
// after successfully validating the access token.
func SetAuthData(c *gin.Context, userID, role string) {
	c.Set(UserIDKey, userID)
	c.Set(RoleKey, role)
}
