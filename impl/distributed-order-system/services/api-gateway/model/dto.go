package model

// CreateOrderRequest is the request body for creating an order.
type CreateOrderRequest struct {
	Products []Product
}

type Product struct {
	ProductId string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

// ProcessPaymentRequest is the request body for processing a payment.
type ProcessPaymentRequest struct {
	Amount float64 `json:"amount"`
}

// APIError is the standard error response format.
type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// HealthResponse is the response for the gateway health endpoint.
type HealthResponse struct {
	RequestID string                       `json:"request_id"`
	Status    string                       `json:"status"`
	Service   string                       `json:"service"`
	Hostname  string                       `json:"hostname"`
	GoVersion string                       `json:"go_version"`
	Uptime    string                       `json:"uptime"`
	Version   string                       `json:"version"`
	CheckedAt string                       `json:"checked_at"`
	GCP       GCPMetadata                  `json:"gcp"`
	Metadata  GatewayMetadata              `json:"metadata"`
	Services  map[string]ServiceHealthInfo `json:"services"`
}

// GCPMetadata holds GCP-specific environment metadata.
type GCPMetadata struct {
	Project     string `json:"project"`
	Region      string `json:"region"`
	ServiceName string `json:"service_name"`
}

// GatewayMetadata holds runtime metadata for the gateway itself.
type GatewayMetadata struct {
	PID        int `json:"pid"`
	Goroutines int `json:"goroutines"`
	CPUCount   int `json:"cpu_count"`
}

// ServiceHealthInfo holds the health check result for a single downstream service.
type ServiceHealthInfo struct {
	Status         string                 `json:"status"`
	StatusCode     int                    `json:"status_code"`
	ResponseTimeMs float64               `json:"response_time_ms"`
	CheckedAt      string                 `json:"checked_at"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Error          string                 `json:"error,omitempty"`
}
