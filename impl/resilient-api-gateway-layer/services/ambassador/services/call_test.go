package services

import (
	"ambassador/configs"
	"ambassador/dtos"
	"ambassador/lib"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	circuitbreaker "ambassador/pkg"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnv struct {
	logger *logrus.Logger
	cbMgr  *circuitbreaker.CircuitBreakerManager
	bhMgr  *circuitbreaker.BulkheadManager
	cs     *CallService
	port   int
}

func setupTestEnv(t *testing.T, handler http.HandlerFunc) *testEnv {
	t.Helper()

	fakeServer := httptest.NewServer(handler)
	t.Cleanup(func() { fakeServer.Close() })

	logger := configs.NewLogger()
	cbMgr := circuitbreaker.NewCircuitBreakerManager()
	bhMgr := circuitbreaker.NewBulkheadManager()

	port := fakeServer.Listener.Addr().(*net.TCPAddr).Port

	lib.Config.Services["test"] = lib.ServiceConfig{
		Host:            "127.0.0.1",
		Port:            port,
		Timeout:         1,
		Threshold:       3,
		MaxConnections:  2,
		QueueTimeout:    200,
		RequestDeadline: 30,
	}

	cb := circuitbreaker.NewCircuitBreaker(3, 1*time.Second)
	cbMgr.Set("test", cb)

	bh := circuitbreaker.NewBulkhead(2, 200*time.Millisecond)
	bhMgr.Set("test", bh)

	cs := NewCallService(logger, cbMgr, bhMgr)

	return &testEnv{
		logger: logger,
		cbMgr:  cbMgr,
		bhMgr:  bhMgr,
		cs:     cs,
		port:   port,
	}
}

func newTestRequest() *dtos.CallRequest {
	return &dtos.CallRequest{
		RequestID:         uuid.New().String(),
		URL:               "/",
		Method:            "GET",
		TargetServiceName: "test",
	}
}

// ─── Test 1: Upstream returns 500s → circuit opens after threshold ───

func TestCallService_CircuitOpensOn500(t *testing.T) {
	env := setupTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	resp, errResp := env.cs.Call(newTestRequest())
	assert.Nil(t, resp)
	require.NotNil(t, errResp)
	assert.Equal(t, dtos.ErrorCodeServiceUnavailable, errResp.ErrorCode)

	cb := env.cbMgr.Get("test")
	assert.Equal(t, circuitbreaker.StateOpen, cb.GetState())

	resp, errResp = env.cs.Call(newTestRequest())
	assert.Nil(t, resp)
	require.NotNil(t, errResp)
	assert.Equal(t, dtos.ErrorCodeServiceUnavailable, errResp.ErrorCode)
}

// ─── Test 2: Upstream times out → verify timeout error + retry behavior ───

func TestCallService_TimeoutTriggersRetry(t *testing.T) {
	var callCount int32

	env := setupTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&callCount, 1)
		time.Sleep(15 * time.Second)
	})

	lib.Config.Services["test"] = lib.ServiceConfig{
		Host:            "127.0.0.1",
		Port:            env.port,
		Timeout:         1,
		Threshold:       3,
		MaxConnections:  5,
		QueueTimeout:    200,
		RequestDeadline: 30,
	}
	bh := circuitbreaker.NewBulkhead(5, 200*time.Millisecond)
	env.bhMgr.Set("test", bh)

	resp, errResp := env.cs.Call(newTestRequest())
	assert.Nil(t, resp)
	require.NotNil(t, errResp)
	assert.Equal(t, dtos.ErrorCodeTimeout, errResp.ErrorCode)

	actualCalls := atomic.LoadInt32(&callCount)
	assert.Equal(t, int32(3), actualCalls, "should have retried threshold times")
}

// ─── Test 3: Upstream recovers → circuit transitions HALF-OPEN → CLOSED ───

func TestCallService_CircuitHalfOpenThenClosedOnRecovery(t *testing.T) {
	var requestCount int32

	env := setupTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		if count <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	// Step 1: First call → 500 → circuit OPEN
	env.cs.Call(newTestRequest())
	cb := env.cbMgr.Get("test")
	assert.Equal(t, circuitbreaker.StateOpen, cb.GetState())

	// Step 2: Wait for cooldown → circuit moves to HALF-OPEN
	time.Sleep(1100 * time.Millisecond)
	assert.True(t, cb.AllowRequest())

	// Step 3: Next call → upstream recovers → circuit CLOSED
	resp, errResp := env.cs.Call(newTestRequest())
	assert.NotNil(t, resp)
	assert.Nil(t, errResp)
	assert.True(t, resp.Success)
	assert.Equal(t, circuitbreaker.StateClosed, cb.GetState())
}

// ─── Test 4: Load/stress test → bulkhead isolation ───

func TestCallService_BulkheadRejectsWhenPoolFull(t *testing.T) {
	env := setupTestEnv(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	lib.Config.Services["test"] = lib.ServiceConfig{
		Host:            "127.0.0.1",
		Port:            env.port,
		Timeout:         1,
		Threshold:       5,
		MaxConnections:  2,
		QueueTimeout:    100,
		RequestDeadline: 10,
	}
	bh := circuitbreaker.NewBulkhead(2, 100*time.Millisecond)
	env.bhMgr.Set("test", bh)

	var wg sync.WaitGroup
	var successCount int32
	var rejectedCount int32

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, errResp := env.cs.Call(newTestRequest())
			if resp != nil && resp.Success {
				atomic.AddInt32(&successCount, 1)
			}
			if errResp != nil && errResp.ErrorCode == dtos.ErrorCodeCapacityExceeded {
				atomic.AddInt32(&rejectedCount, 1)
			}
		}()
	}

	wg.Wait()

	fmt.Printf("success=%d rejected=%d\n", successCount, rejectedCount)
	assert.Equal(t, int32(2), successCount, "only 2 should succeed (max_connections=2)")
	assert.True(t, rejectedCount >= 1, "at least 1 should be rejected by bulkhead")
}
