package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shared/middleware"
	"shared/response"

	"web-backend/usecases/auth"
)

// TokenResponseStrategy defines how tokens are delivered to the client.
type TokenResponseStrategy interface {
	SendTokens(c *gin.Context, accessToken, refreshToken string) error
	ClearTokens(c *gin.Context) error
}

// CookieTokenStrategy implements TokenResponseStrategy using HTTP-only cookies.
type CookieTokenStrategy struct{}

// NewCookieTokenStrategy creates a new cookie-based token strategy.
func NewCookieTokenStrategy() *CookieTokenStrategy {
	return &CookieTokenStrategy{}
}

// SendTokens stores access and refresh tokens in HTTP-only cookies.
func (s *CookieTokenStrategy) SendTokens(c *gin.Context, accessToken, refreshToken string) error {
	c.SetCookie("access_token", accessToken, 900, "/", "", false, true)
	c.SetCookie("refresh_token", refreshToken, 604800, "/", "", false, true)
	return nil
}

// ClearTokens removes the tokens from cookies.
func (s *CookieTokenStrategy) ClearTokens(c *gin.Context) error {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
	return nil
}

// AuthController defines the interface for auth HTTP handlers.
type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
}

type authController struct {
	usecase  auth.AuthUseCase
	strategy TokenResponseStrategy
}

// NewAuthController creates a new AuthController.
func NewAuthController(usecase auth.AuthUseCase, strategy TokenResponseStrategy) AuthController {
	return &authController{usecase: usecase, strategy: strategy}
}

// Register handles POST /auth/register.
func (ctrl *authController) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := ctrl.usecase.Register(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusCreated, "user registered", resp)
}

// Login handles POST /auth/login.
func (ctrl *authController) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := ctrl.usecase.Login(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	if err := ctrl.strategy.SendTokens(c, resp.AccessToken, resp.RefreshToken); err != nil {
		response.InternalError(c, "failed to send tokens")
		return
	}
}

// Refresh handles POST /auth/refresh.
func (ctrl *authController) Refresh(c *gin.Context) {
	// Try to get refresh token from cookie first (web), then from body (fallback).
	refreshToken := ""
	if cookie, err := c.Cookie("refresh_token"); err == nil {
		refreshToken = cookie
	}

	if refreshToken == "" {
		var req auth.RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
			response.BadRequest(c, "refresh token is required")
			return
		}
		refreshToken = req.RefreshToken
	}

	resp, err := ctrl.usecase.Refresh(c.Request.Context(), auth.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	if err := ctrl.strategy.SendTokens(c, resp.AccessToken, resp.RefreshToken); err != nil {
		response.InternalError(c, "failed to send tokens")
		return
	}
}

// Logout handles POST /auth/logout.
func (ctrl *authController) Logout(c *gin.Context) {
	if err := ctrl.strategy.ClearTokens(c); err != nil {
		response.InternalError(c, "failed to clear tokens")
		return
	}
	response.Success(c, http.StatusOK, "logged out", nil)
}

// GetUserID extracts the authenticated user ID from the Gin context.
// This is a convenience wrapper for controllers.
func GetUserID(c *gin.Context) string {
	return middleware.GetUserID(c)
}
