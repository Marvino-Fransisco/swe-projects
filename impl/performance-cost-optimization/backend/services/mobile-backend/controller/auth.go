package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shared/response"

	"mobile-backend/usecases/auth"
)

// AuthController defines the interface for auth HTTP handlers.
type AuthController interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Refresh(c *gin.Context)
	Logout(c *gin.Context)
}

type authController struct {
	usecase auth.AuthUseCase
}

// NewAuthController creates a new AuthController.
func NewAuthController(usecase auth.AuthUseCase) AuthController {
	return &authController{usecase: usecase}
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
// Returns access and refresh tokens in the JSON response body.
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

	response.Success(c, http.StatusOK, "authentication successful", resp)
}

// Refresh handles POST /auth/refresh.
// Returns new access and refresh tokens in the JSON response body.
func (ctrl *authController) Refresh(c *gin.Context) {
	var req auth.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "refresh token is required")
		return
	}

	resp, err := ctrl.usecase.Refresh(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "tokens refreshed", resp)
}

// Logout handles POST /auth/logout.
// The mobile client is responsible for discarding the stored tokens.
func (ctrl *authController) Logout(c *gin.Context) {
	response.Success(c, http.StatusOK, "logged out successfully", nil)
}
