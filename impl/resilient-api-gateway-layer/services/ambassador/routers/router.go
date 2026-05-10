package routers

import (
	"ambassador/controllers"
	"ambassador/dtos"
	"net/http"

	"ambassador/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func SetupRouter(logger *logrus.Logger, callController *controllers.CallController, healthController *controllers.HealthController) *gin.Engine {
	r := gin.Default()

	r.Use(middlewares.CreateRequestID(logger))

	r.POST("/call", callController.Call)
	r.GET("/status", healthController.Status)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, dtos.ErrorResponse{
			Success:   false,
			ErrorCode: dtos.ErrorCodeResourceNotFound,
			Message:   "route not found",
		})
	})

	return r
}
