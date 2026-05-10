package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_StartsClosed(t *testing.T) {
	cb := NewCircuitBreaker(5, 10*time.Second)
	assert.Equal(t, "closed", cb.GetState())
	assert.True(t, cb.AllowRequest())
}
func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, 10*time.Second)
	cb.RecordFailure() // count=1
	cb.RecordFailure() // count=2
	cb.RecordFailure() // count=3 → threshold hit → OPEN
	assert.Equal(t, "open", cb.GetState())
	assert.False(t, cb.AllowRequest())
}
func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(1, 1*time.Nanosecond) // instant timeout
	cb.RecordFailure()                            // → OPEN
	time.Sleep(10 * time.Millisecond)             // wait for cooldown
	assert.True(t, cb.AllowRequest())             // → HALF-OPEN, allows probe
	assert.Equal(t, "half-open", cb.GetState())
}
func TestCircuitBreaker_ClosesAfterSuccessInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(1, 1*time.Nanosecond)
	cb.RecordFailure() // → OPEN
	time.Sleep(10 * time.Millisecond)
	cb.AllowRequest()  // → HALF-OPEN
	cb.RecordSuccess() // → CLOSED
	assert.Equal(t, "closed", cb.GetState())
}
