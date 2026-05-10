package services

import (
	"ambassador/dtos"
	"ambassador/lib"
	circuitbreaker "ambassador/pkg"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var errTCPConnectionTimeout = errors.New("tcp connection timeout")

type CallService struct {
	log   *logrus.Entry
	cbMgr *circuitbreaker.CircuitBreakerManager
	bhMgr *circuitbreaker.BulkheadManager
}

func NewCallService(logger *logrus.Logger, cbMgr *circuitbreaker.CircuitBreakerManager, bhMgr *circuitbreaker.BulkheadManager) *CallService {
	log := logger.WithField("service", "CallService")
	return &CallService{
		log:   log,
		cbMgr: cbMgr,
		bhMgr: bhMgr,
	}
}

func (s *CallService) Call(req *dtos.CallRequest) (*dtos.CallResponse, *dtos.ErrorResponse) {
	normalizeRequest(req)

	requestID := req.RequestID
	startTime := time.Now()
	log := s.log.WithFields(logrus.Fields{
		"function":       "call",
		"request_id":     requestID,
		"target_service": req.TargetServiceName,
		"method":         req.Method,
		"path":           req.URL,
	})

	var retryCount int
	var statusCode int
	var finalState string

	defer func() {
		latency := time.Since(startTime)

		bh := s.bhMgr.Get(req.TargetServiceName)
		bulkheadActive := 0
		bulkheadMax := 0
		if bh != nil {
			bulkheadActive = bh.ActiveConnections()
			bulkheadMax = bh.MaxConnections()
		}

		cb := s.cbMgr.Get(req.TargetServiceName)
		if cb != nil {
			finalState = cb.GetState()
		}

		log.WithFields(logrus.Fields{
			"status_code":      statusCode,
			"retry_count":      retryCount,
			"circuit_state":    finalState,
			"bulkhead_active":  bulkheadActive,
			"bulkhead_max":     bulkheadMax,
			"total_latency_ms": latency.Milliseconds(),
		}).Info("Request completed")
	}()

	// Acquire bulkhead slot first — fail fast if connection pool is full
	bh := s.bhMgr.Get(req.TargetServiceName)
	if bh != nil {
		if err := bh.Acquire(); err != nil {
			log.Warnf("Bulkhead full for service: %s, rejecting request", req.TargetServiceName)
			statusCode = http.StatusServiceUnavailable
			return nil, &dtos.ErrorResponse{
				Success:   false,
				ErrorCode: dtos.ErrorCodeCapacityExceeded,
				Message:   fmt.Sprintf("max concurrent connections reached for service: %s", req.TargetServiceName),
				RequestID: requestID,
			}
		}
		defer bh.Release()
	}

	// Get the circuit breaker for the target service
	cb := s.cbMgr.Get(req.TargetServiceName)
	if cb == nil {
		log.Errorf("No circuit breaker found for service: %s", req.TargetServiceName)
		statusCode = http.StatusInternalServerError
		return nil, newError(requestID, dtos.ErrorCodeInternalServerError, fmt.Sprintf("no circuit breaker configured for service: %s", req.TargetServiceName))
	}

	// Check circuit breaker state
	if !cb.AllowRequest() {
		log.Warnf("Circuit breaker is OPEN for service: %s, failing fast", req.TargetServiceName)
		statusCode = http.StatusServiceUnavailable
		return nil, newError(requestID, dtos.ErrorCodeServiceUnavailable, fmt.Sprintf("circuit breaker is open for service: %s", req.TargetServiceName))
	}

	// Look up service config by TargetServiceName
	serviceConfig, exists := lib.Config.Services[req.TargetServiceName]
	if !exists {
		log.Errorf("No service config found for service: %s", req.TargetServiceName)
		statusCode = http.StatusInternalServerError
		return nil, newError(requestID, dtos.ErrorCodeInternalServerError, fmt.Sprintf("no service config found for service: %s", req.TargetServiceName))
	}

	// Build URL from config host:port + request path
	targetURL := fmt.Sprintf("http://%s:%d%s", serviceConfig.Host, serviceConfig.Port, req.URL)

	header := &dtos.RequestHeader{
		XRequestID:   requestID,
		XServiceName: req.TargetServiceName,
	}

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(req.Body)
		if err != nil {
			log.Errorf("Failed to marshal request body: %v", err)
			statusCode = http.StatusInternalServerError
			return nil, newError(requestID, dtos.ErrorCodeInternalServerError, "failed to process request body")
		}
	}

	log.Infof("Calling upstream service: %s %s", req.Method, targetURL)

	// Set overall request deadline (wall-clock limit including all retries)
	deadline := time.Now().Add(time.Duration(serviceConfig.RequestDeadline) * time.Second)

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, addr)
			if err != nil {
				if ctx.Err() == context.DeadlineExceeded || strings.Contains(err.Error(), "timeout") {
					return nil, errTCPConnectionTimeout
				}
				return nil, err
			}
			return conn, nil
		},
		ResponseHeaderTimeout: 10 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
	}

	maxRetries := cb.GetFailureThreshold()
	var lastErrResp *dtos.ErrorResponse

	for i := 0; i < maxRetries; i++ {
		retryCount = i + 1

		// Check if the overall request deadline has been exceeded
		if time.Now().After(deadline) {
			log.Warnf("Request deadline exceeded for service: %s after %d attempts", req.TargetServiceName, i)
			statusCode = http.StatusGatewayTimeout
			return nil, newError(requestID, dtos.ErrorCodeTimeout, "request deadline exceeded")
		}

		// Before each attempt, check circuit breaker
		if !cb.AllowRequest() {
			log.Warnf("Circuit breaker is OPEN for service: %s on attempt %d, failing fast", req.TargetServiceName, i+1)
			statusCode = http.StatusServiceUnavailable
			return nil, newError(requestID, dtos.ErrorCodeServiceUnavailable, fmt.Sprintf("circuit breaker is open for service: %s", req.TargetServiceName))
		}

		// Re-create body reader for each attempt since bytes.Reader is consumed
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		request, err := http.NewRequest(req.Method, targetURL, bodyReader)
		if err != nil {
			log.Errorf("Failed to create http request: %v", err)
			statusCode = http.StatusInternalServerError
			return nil, newError(requestID, dtos.ErrorCodeInternalServerError, "failed to create http request")
		}

		setRequestHeader(header, request, deadline)

		resp, err := client.Do(request)
		if err != nil {
			log.Errorf("Failed to execute http request (attempt %d/%d): %v", i+1, maxRetries, err)

			cb.RecordFailure()

			// Check if it's a timeout error
			if errors.Is(err, errTCPConnectionTimeout) {
				log.Warnf("TCP Connection Timeout on attempt %d, aborting retry, %s", i+1, req.Method)
				statusCode = http.StatusGatewayTimeout
				lastErrResp = normalizeRequestError(err)
				break
			}

			if isTimeoutError(err) {
				if !isIdempotent(req.Method) {
					log.Warnf("Timeout on attempt %d for non-idempotent method %s, aborting retry", i+1, req.Method)
					statusCode = http.StatusGatewayTimeout
					lastErrResp = normalizeRequestError(err)
					break
				}
				sleepDur := setTimeSleep(i)
				log.Warnf("Connection timeout on attempt %d, waiting %v before retry", i+1, sleepDur)
				lastErrResp = normalizeRequestError(err)
				time.Sleep(sleepDur)
				continue
			}

			sleepDur := setTimeSleep(i)
			log.Warnf("Connection error on attempt %d, retrying with backoff %v", i+1, sleepDur)
			statusCode = http.StatusBadGateway
			lastErrResp = normalizeRequestError(err)
			time.Sleep(sleepDur)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Errorf("Failed to read response body: %v", err)
			statusCode = http.StatusInternalServerError
			return nil, newError(requestID, dtos.ErrorCodeInternalServerError, "failed to read response from upstream")
		}

		log.Infof("Response status: %d, body: %s", resp.StatusCode, string(respBody))

		// Handle HTTP error responses
		if resp.StatusCode >= 400 {
			statusCode = resp.StatusCode
			log.Errorf("Upstream returned status %d, body: %s", resp.StatusCode, string(respBody))

			// 5xx or auth errors (401, 403) — open circuit breaker immediately, no retry
			if resp.StatusCode >= 500 || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
				cb.RecordFailure()
				cb.SetState(circuitbreaker.StateOpen)
				return nil, newError(requestID, normalizeHTTPErrorCode(resp.StatusCode), sanitizeHTTPErrorMessage(resp.StatusCode))
			}

			// Other 4xx (400, 404) — return error immediately, do NOT open circuit breaker
			return nil, newError(requestID, normalizeHTTPErrorCode(resp.StatusCode), sanitizeHTTPErrorMessage(resp.StatusCode))
		}

		// Success
		cb.RecordSuccess()
		statusCode = resp.StatusCode

		var data any
		if err := json.Unmarshal(respBody, &data); err != nil {
			data = string(respBody)
		}

		return &dtos.CallResponse{
			Success: true,
			Data:    data,
		}, nil
	}

	// All retries exhausted — open the circuit breaker
	retryCount = maxRetries
	statusCode = http.StatusServiceUnavailable
	log.Errorf("All %d retries exhausted for service: %s", maxRetries, req.TargetServiceName)
	cb.SetState(circuitbreaker.StateOpen)
	if lastErrResp != nil {
		lastErrResp.RequestID = requestID
		return nil, lastErrResp
	}
	return nil, newError(requestID, dtos.ErrorCodeMaxRetriesExceeded, fmt.Sprintf("all retries exhausted for service: %s", req.TargetServiceName))
}

func isTimeoutError(err error) bool {
	if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
		return true
	}
	return strings.Contains(err.Error(), "connection refused") == false &&
		strings.Contains(err.Error(), "no such host") == false &&
		(strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded"))
}

func normalizeRequestError(err error) *dtos.ErrorResponse {
	if errors.Is(err, errTCPConnectionTimeout) {
		return &dtos.ErrorResponse{
			Success:   false,
			ErrorCode: dtos.ErrorTCPConnectionTimeout,
			Message:   "failed to establish connection to downstream service",
		}
	}
	if err, ok := err.(interface{ Timeout() bool }); ok && err.Timeout() {
		return &dtos.ErrorResponse{
			Success:   false,
			ErrorCode: dtos.ErrorCodeTimeout,
			Message:   "request to downstream service timed out",
		}
	}
	return &dtos.ErrorResponse{
		Success:   false,
		ErrorCode: dtos.ErrorCodeServiceUnavailable,
		Message:   "failed to reach downstream service",
	}
}

func newError(requestID string, code dtos.ErrorCode, message string) *dtos.ErrorResponse {
	return &dtos.ErrorResponse{
		Success:   false,
		ErrorCode: code,
		Message:   message,
		RequestID: requestID,
	}
}

func normalizeHTTPErrorCode(statusCode int) dtos.ErrorCode {
	switch {
	case statusCode == http.StatusNotFound:
		return dtos.ErrorCodeResourceNotFound
	case statusCode == http.StatusUnauthorized:
		return dtos.ErrorCodeUnauthorized
	case statusCode == http.StatusForbidden:
		return dtos.ErrorCodeForbidden
	case statusCode == http.StatusBadRequest:
		return dtos.ErrorCodeBadRequest
	case statusCode >= 500:
		return dtos.ErrorCodeServiceUnavailable
	default:
		return dtos.ErrorCodeInternalServerError
	}
}

func sanitizeHTTPErrorMessage(statusCode int) string {
	switch {
	case statusCode == http.StatusNotFound:
		return "requested resource not found"
	case statusCode == http.StatusUnauthorized:
		return "authentication required"
	case statusCode == http.StatusForbidden:
		return "access denied"
	case statusCode == http.StatusBadRequest:
		return "invalid request"
	case statusCode >= 500:
		return "upstream service error"
	default:
		return fmt.Sprintf("unexpected response from upstream (status %d)", statusCode)
	}
}

func setRequestHeader(header *dtos.RequestHeader, req *http.Request, deadline time.Time) {
	req.Header.Set("X-Request-ID", header.XRequestID)
	req.Header.Set("X-Service-Name", header.XServiceName)

	remaining := time.Until(deadline)
	if remaining > 0 {
		req.Header.Set("X-Request-Timeout", fmt.Sprintf("%dms", remaining.Milliseconds()))
	}
}

func normalizeRequest(req *dtos.CallRequest) {
	req.Method = strings.ToUpper(req.Method)
}

func isIdempotent(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func setTimeSleep(attempt int) time.Duration {
	cap := 30
	base := 1

	backoff := min(cap, base<<attempt)
	half := backoff / 2
	wait := half + rand.Intn(half)

	return time.Duration(wait) * time.Second
}
