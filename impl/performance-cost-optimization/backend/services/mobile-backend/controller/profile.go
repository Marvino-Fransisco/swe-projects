package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"shared/middleware"
	"shared/response"

	"mobile-backend/usecases/profile"
)

// ProfileController defines the interface for profile HTTP handlers.
type ProfileController interface {
	GetProfile(c *gin.Context)
	UpdateProfile(c *gin.Context)
	ChangePassword(c *gin.Context)
}

type profileController struct {
	usecase profile.ProfileUseCase
}

// NewProfileController creates a new ProfileController.
func NewProfileController(usecase profile.ProfileUseCase) ProfileController {
	return &profileController{usecase: usecase}
}

// GetProfile handles GET /profile.
func (ctrl *profileController) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	resp, err := ctrl.usecase.GetProfile(c.Request.Context(), profile.GetProfileRequest{
		UserID: userID,
	})
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "profile retrieved", resp)
}

// UpdateProfile handles PUT /profile.
func (ctrl *profileController) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req profile.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.UserID = userID

	resp, err := ctrl.usecase.UpdateProfile(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "profile updated", resp)
}

// ChangePassword handles PUT /profile/password.
func (ctrl *profileController) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req profile.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.UserID = userID

	if err := ctrl.usecase.ChangePassword(c.Request.Context(), req); err != nil {
		handleError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "password changed", nil)
}
