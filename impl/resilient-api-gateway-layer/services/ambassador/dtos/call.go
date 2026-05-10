package dtos

type CallRequest struct {
	RequestID         string
	URL               string            `json:"url" binding:"required"`
	Method            string            `json:"method" binding:"required"`
	Headers           map[string]string `json:"headers"`
	Body              any               `json:"body"`
	TargetServiceName string            `json:"target_service_name" binding:"required"`
}

type CallResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data,omitempty"`
}
