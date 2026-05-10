package configs

import (
	"ambassador/lib"

	"github.com/sirupsen/logrus"
)

func NewLogger() *logrus.Logger {
	if lib.Env.Environment == "dev" {
		return InitDevLogger()
	} else {
		return InitProdLogger()
	}
}

func InitDevLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(true)

	return logger
}

func InitProdLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
