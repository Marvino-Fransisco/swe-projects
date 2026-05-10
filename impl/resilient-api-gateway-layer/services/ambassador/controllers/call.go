package controllers

import (
	"net/http"

	"ambassador/dtos"
	"ambassador/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type CallController struct {
	callService *services.CallService
	log         *logrus.Entry
}

func NewCallController(callService *services.CallService, logger *logrus.Logger) *CallController {
	log := logger.WithField("controller", "CallController")

	return &CallController{
		callService: callService,
		log:         log,
	}
}

func (ctrl *CallController) Call(c *gin.Context) {
	requestLog := c.MustGet("log").(*logrus.Entry)
	log := requestLog.WithFields(ctrl.log.Data).WithField("function", "Call")

	var req dtos.CallRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, dtos.ErrorResponse{
			Success:   false,
			ErrorCode: dtos.ErrorCodeBadRequest,
			Message:   "invalid request body: " + err.Error(),
		})
		return
	}

	resp, errResp := ctrl.callService.Call(&req)
	if errResp != nil {
		log.WithField("error_code", errResp.ErrorCode).Error(errResp.Message)
		c.JSON(errorCodeToHTTPStatus(errResp.ErrorCode), errResp)
		return
	}

	log.Info("Call Success")
	c.JSON(http.StatusOK, resp)
}

func errorCodeToHTTPStatus(code dtos.ErrorCode) int {
	switch code {
	case dtos.ErrorCodeBadRequest:
		return http.StatusBadRequest
	case dtos.ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case dtos.ErrorCodeForbidden:
		return http.StatusForbidden
	case dtos.ErrorCodeResourceNotFound:
		return http.StatusNotFound
	case dtos.ErrorCodeTimeout:
		return http.StatusGatewayTimeout
	case dtos.ErrorCodeServiceUnavailable:
		return http.StatusBadGateway
	case dtos.ErrorCodeInternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
