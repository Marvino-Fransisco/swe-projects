package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ServiceClient is an HTTP client for calling downstream services.
type ServiceClient struct {
	httpClient *http.Client
}

// Response wraps the downstream service HTTP response.
type Response struct {
	StatusCode int
	Body       []byte
}

// NewServiceClient creates a new ServiceClient with a timeout.
func NewServiceClient() *ServiceClient {
	return &ServiceClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Get sends a GET request to the given URL and returns the response.
func (c *ServiceClient) Get(url string) (*Response, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}

// Post sends a POST request with a JSON payload and returns the response.
func (c *ServiceClient) Post(url string, payload interface{}) (*Response, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to call service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}
