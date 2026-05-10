package middleware

import (
	"github.com/gin-gonic/gin"

	sharedMiddleware "shared/middleware"
	"shared/response"
	"shared/util"
)

// authMiddleware implements the sharedMiddleware.AuthMiddleware interface
// using HTTP-only cookies for the website BFF.
type authMiddleware struct {
	jwtSvc *util.JWTService
}

// NewAuthMiddleware creates a new cookie-based auth middleware.
func NewAuthMiddleware(jwtSvc *util.JWTService) sharedMiddleware.AuthMiddleware {
	return &authMiddleware{
		jwtSvc: jwtSvc,
	}
}

// RequireAuth validates the access token from the HTTP-only cookie.
// Aborts with 401 if the token is missing or invalid.
func (m *authMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			response.Unauthorized(c, "missing access token cookie")
			c.Abort()
			return
		}

		claims, err := m.jwtSvc.ValidateAccessToken(accessToken)
		if err != nil {
			response.Unauthorized(c, "invalid or expired access token")
			c.Abort()
			return
		}

		sharedMiddleware.SetAuthData(c, claims.UserID, claims.Role)
		c.Next()
	}
}

// OptionalAuth attempts to validate the access token from the cookie
// but does not abort if it is missing or invalid.
func (m *authMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("access_token")
		if err != nil {
			c.Next()
			return
		}

		claims, err := m.jwtSvc.ValidateAccessToken(accessToken)
		if err != nil {
			c.Next()
			return
		}

		sharedMiddleware.SetAuthData(c, claims.UserID, claims.Role)
		c.Next()
	}
}
