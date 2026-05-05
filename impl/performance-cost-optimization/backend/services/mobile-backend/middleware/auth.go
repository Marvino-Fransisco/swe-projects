package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"shared/middleware"
	"shared/response"
	"shared/util"
)

// authMiddleware implements the shared middleware.AuthMiddleware interface
// using the Authorization header (Bearer token) for the mobile BFF.
type authMiddleware struct {
	jwtSvc *util.JWTService
}

// NewAuthMiddleware creates a new bearer-token-based auth middleware.
func NewAuthMiddleware(jwtSvc *util.JWTService) middleware.AuthMiddleware {
	return &authMiddleware{
		jwtSvc: jwtSvc,
	}
}

// RequireAuth validates the access token from the Authorization header.
// Aborts with 401 if the token is missing or invalid.
func (m *authMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		claims, err := m.jwtSvc.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "invalid or expired access token")
			c.Abort()
			return
		}

		middleware.SetAuthData(c, claims.UserID, claims.Role)
		c.Next()
	}
}

// OptionalAuth attempts to validate the access token from the Authorization header
// but does not abort if it is missing or invalid.
func (m *authMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwtSvc.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		middleware.SetAuthData(c, claims.UserID, claims.Role)
		c.Next()
	}
}

// extractBearerToken extracts the bearer token from the Authorization header.
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
