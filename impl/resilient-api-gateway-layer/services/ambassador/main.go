package main

import (
	"ambassador/configs"
	"ambassador/controllers"
	"ambassador/lib"
	circuitbreaker "ambassador/pkg"
	"ambassador/routers"
	"ambassador/services"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := configs.NewLogger()

	// Initialize circuit breaker manager
	cbMgr := circuitbreaker.NewCircuitBreakerManager()

	// Initialize bulkhead manager
	bhMgr := circuitbreaker.NewBulkheadManager()

	// Create circuit breaker and bulkhead for each configured service
	for serviceName, serviceConfig := range lib.Config.Services {
		cb := circuitbreaker.NewCircuitBreaker(
			serviceConfig.Threshold,
			time.Duration(serviceConfig.Timeout)*time.Second,
		)
		cbMgr.Set(serviceName, cb)

		bh := circuitbreaker.NewBulkhead(
			serviceConfig.MaxConnections,
			time.Duration(serviceConfig.QueueTimeout)*time.Millisecond,
		)
		bhMgr.Set(serviceName, bh)
	}

	callService := services.NewCallService(logger, cbMgr, bhMgr)
	callController := controllers.NewCallController(callService, logger)

	healthService := services.NewHealthService(logger, cbMgr, bhMgr)
	healthController := controllers.NewHealthController(healthService, logger)

	r := routers.SetupRouter(logger, callController, healthController)

	log := logger.WithFields(logrus.Fields{
		"env": lib.Env.Environment,
	})

	log.Info("Server is running in port ", lib.Env.Port)
	if err := r.Run(":6969"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
