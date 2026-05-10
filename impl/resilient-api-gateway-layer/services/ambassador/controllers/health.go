package controllers

import (
	"ambassador/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type HealthController struct {
	healthService *services.HealthService
	log           *logrus.Entry
}

func NewHealthController(healthService *services.HealthService, logger *logrus.Logger) *HealthController {
	log := logger.WithField("controller", "HealthController")

	return &HealthController{
		healthService: healthService,
		log:           log,
	}
}

func (ctrl *HealthController) Status(c *gin.Context) {
	requestLog := c.MustGet("log").(*logrus.Entry)
	log := requestLog.WithFields(ctrl.log.Data).WithField("function", "Status")

	resp := ctrl.healthService.GetStatus()

	log.Info("Health status requested")
	c.JSON(http.StatusOK, resp)
}
