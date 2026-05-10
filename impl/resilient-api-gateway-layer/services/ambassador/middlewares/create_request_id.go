package middlewares

import (
	"ambassador/lib"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func CreateRequestID(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		log := logger.WithFields(logrus.Fields{
			"requestID": requestID,
			"env":       lib.Env.Environment,
		})

		c.Set("requestID", requestID)
		c.Set("log", log)
		c.Next()
	}
}
