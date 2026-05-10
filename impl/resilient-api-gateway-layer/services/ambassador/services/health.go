package services

import (
	"ambassador/dtos"
	circuitbreaker "ambassador/pkg"

	"github.com/sirupsen/logrus"
)

type HealthService struct {
	log   *logrus.Entry
	cbMgr *circuitbreaker.CircuitBreakerManager
	bhMgr *circuitbreaker.BulkheadManager
}

func NewHealthService(logger *logrus.Logger, cbMgr *circuitbreaker.CircuitBreakerManager, bhMgr *circuitbreaker.BulkheadManager) *HealthService {
	log := logger.WithField("service", "HealthService")
	return &HealthService{
		log:   log,
		cbMgr: cbMgr,
		bhMgr: bhMgr,
	}
}

func (s *HealthService) GetStatus() *dtos.HealthResponse {
	log := s.log.WithField("function", "GetStatus")

	// Gather circuit breaker states
	cbAll := s.cbMgr.GetAll()
	cbStates := make([]dtos.CircuitBreakerState, 0, len(cbAll))

	for serviceName, cb := range cbAll {
		cbStates = append(cbStates, dtos.CircuitBreakerState{
			ServiceName:      serviceName,
			State:            cb.GetState(),
			FailureCount:     cb.GetFailureCount(),
			FailureThreshold: cb.GetFailureThreshold(),
		})
	}

	// Gather bulkhead states
	bhAll := s.bhMgr.GetAll()
	bhStates := make([]dtos.BulkheadState, 0, len(bhAll))

	for serviceName, bh := range bhAll {
		bhStates = append(bhStates, dtos.BulkheadState{
			ServiceName:       serviceName,
			ActiveConnections: bh.ActiveConnections(),
			MaxConnections:    bh.MaxConnections(),
		})
	}

	log.Infof("Returning health status for %d services", len(cbStates))

	return &dtos.HealthResponse{
		Success:         true,
		CircuitBreakers: cbStates,
		Bulkheads:       bhStates,
	}
}
