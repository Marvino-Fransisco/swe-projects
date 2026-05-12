package handler

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"api-gateway/client"
	"api-gateway/model"

	"github.com/gin-gonic/gin"
)

var startTime = time.Now()

// downstreamService defines a service to health-check.
type downstreamService struct {
	Name string
	URL  string
}

// HealthHandler handles health check requests for the API gateway.
type HealthHandler struct {
	client     *client.ServiceClient
	services   []downstreamService
}

// NewHealthHandler creates a new HealthHandler with the downstream service URLs.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		client: client.NewServiceClient(),
		services: []downstreamService{
			{Name: "inventory-service", URL: "http://inventory_devcontainer-app-1:8001/health"},
			{Name: "order-service", URL: "http://order_devcontainer-app-1:8002/health"},
			{Name: "payment-service", URL: "http://payment_devcontainer-app-1:8003/health"},
		},
	}
}

// Check handles GET /api/health
// It checks the health of the gateway itself and all downstream services concurrently.
func (h *HealthHandler) Check(c *gin.Context) {
	requestID := generateUUID()
	checkedAt := time.Now().UTC().Format(time.RFC3339)
	hostname, _ := os.Hostname()

	// Check all downstream services concurrently.
	servicesHealth := h.checkDownstreamServices()

	response := model.HealthResponse{
		RequestID: requestID,
		Status:    "ok",
		Service:   "api-gateway",
		Hostname:  hostname,
		GoVersion: runtime.Version(),
		Uptime:    time.Since(startTime).String(),
		Version:   os.Getenv("VERSION"),
		CheckedAt: checkedAt,
		GCP: model.GCPMetadata{
			Project:     os.Getenv("GCP_PROJECT"),
			Region:      os.Getenv("GCP_REGION"),
			ServiceName: os.Getenv("K_SERVICE"),
		},
		Metadata: model.GatewayMetadata{
			PID:        os.Getpid(),
			Goroutines: runtime.NumGoroutine(),
			CPUCount:   runtime.NumCPU(),
		},
		Services: servicesHealth,
	}

	c.JSON(http.StatusOK, response)
}

// checkDownstreamServices calls each downstream service's /health endpoint concurrently
// and returns a map of service name to health info.
func (h *HealthHandler) checkDownstreamServices() map[string]model.ServiceHealthInfo {
	results := make(map[string]model.ServiceHealthInfo, len(h.services))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, svc := range h.services {
		wg.Add(1)
		go func(s downstreamService) {
			defer wg.Done()
			info := h.checkSingleService(s)

			mu.Lock()
			results[s.Name] = info
			mu.Unlock()
		}(svc)
	}

	wg.Wait()
	return results
}

// checkSingleService performs a health check against a single downstream service
// and returns the result including response time and status.
func (h *HealthHandler) checkSingleService(svc downstreamService) model.ServiceHealthInfo {
	start := time.Now()
	checkedAt := time.Now().UTC().Format(time.RFC3339)

	resp, err := h.client.Get(svc.URL)
	elapsed := time.Since(start)
	responseTimeMs := float64(elapsed.Nanoseconds()) / 1e6

	if err != nil {
		return model.ServiceHealthInfo{
			Status:         "error",
			StatusCode:     0,
			ResponseTimeMs: responseTimeMs,
			CheckedAt:      checkedAt,
			Error:          err.Error(),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return model.ServiceHealthInfo{
			Status:         "error",
			StatusCode:     resp.StatusCode,
			ResponseTimeMs: responseTimeMs,
			CheckedAt:      checkedAt,
			Error:          string(resp.Body),
		}
	}

	// Parse the downstream service's health response into a map.
	var data map[string]interface{}
	if unmarshalErr := json.Unmarshal(resp.Body, &data); unmarshalErr != nil {
		return model.ServiceHealthInfo{
			Status:         "error",
			StatusCode:     resp.StatusCode,
			ResponseTimeMs: responseTimeMs,
			CheckedAt:      checkedAt,
			Error:          fmt.Sprintf("failed to parse response: %s", unmarshalErr.Error()),
		}
	}

	return model.ServiceHealthInfo{
		Status:         "ok",
		StatusCode:     resp.StatusCode,
		ResponseTimeMs: responseTimeMs,
		CheckedAt:      checkedAt,
		Data:           data,
	}
}

// generateUUID creates a version 4 UUID using crypto/rand.
func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
