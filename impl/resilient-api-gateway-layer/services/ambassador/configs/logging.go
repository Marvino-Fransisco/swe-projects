package configs

import (
	"ambassador/lib"

	"github.com/sirupsen/logrus"
)

func NewLogger() *logrus.Logger {
	if lib.Env.Environment == "dev" || lib.Env.Environment == "development" {
		return InitDevLogger()
	}
	return InitProdLogger()
}

func InitDevLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	logger.SetLevel(logrus.InfoLevel)

	return logger
}

func InitProdLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	return logger
}
