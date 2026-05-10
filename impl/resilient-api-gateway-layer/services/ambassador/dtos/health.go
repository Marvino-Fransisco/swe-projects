package dtos

type CircuitBreakerState struct {
	ServiceName      string `json:"service_name"`
	State            string `json:"state"`
	FailureCount     int    `json:"failure_count"`
	FailureThreshold int    `json:"failure_threshold"`
}

type BulkheadState struct {
	ServiceName       string `json:"service_name"`
	ActiveConnections int    `json:"active_connections"`
	MaxConnections    int    `json:"max_connections"`
}

type HealthResponse struct {
	Success         bool                  `json:"success"`
	CircuitBreakers []CircuitBreakerState `json:"circuit_breakers"`
	Bulkheads       []BulkheadState       `json:"bulkheads"`
}
