package pkg

import (
	"sync"
	"time"
)

var (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half-open"
)

var cbTransitions = []Transition{
	{From: StateClosed, Event: EventSuccess, To: StateClosed},
	{From: StateClosed, Event: EventFailure, To: StateClosed},
	{From: StateClosed, Event: EventThresholdHit, To: StateOpen},
	{From: StateClosed, Event: EventForceOpen, To: StateOpen},

	{From: StateOpen, Event: EventCooldownTimer, To: StateHalfOpen},
	{From: StateOpen, Event: EventForceOpen, To: StateOpen},

	{From: StateHalfOpen, Event: EventSuccess, To: StateClosed},
	{From: StateHalfOpen, Event: EventFailure, To: StateOpen},
	{From: StateHalfOpen, Event: EventForceOpen, To: StateOpen},
}

type CircuitBreaker struct {
	mu               sync.Mutex
	fsm              *FSM
	failureCount     int
	failureThreshold int
	lastFailureTime  time.Time
	timeout          time.Duration
}

func NewCircuitBreaker(failureThreshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		fsm:              NewFSM(StateClosed, cbTransitions),
		failureCount:     0,
		failureThreshold: failureThreshold,
		timeout:          timeout,
	}
}

func (cb *CircuitBreaker) SetState(state string) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if state == StateOpen {
		cb.fsm.Transition(EventForceOpen)
		return
	}

	cb.fsm.current = state
}

func (cb *CircuitBreaker) GetState() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.fsm.Current()
}

func (cb *CircuitBreaker) AllowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.fsm.Current()

	if state == StateClosed {
		return true
	}

	if state == StateOpen {
		if time.Since(cb.lastFailureTime) >= cb.timeout {
			cb.fsm.Transition(EventCooldownTimer)
			return true
		}
		return false
	}

	return true
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.fsm.Transition(EventSuccess)
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	state := cb.fsm.Current()

	if state == StateHalfOpen {
		cb.fsm.Transition(EventFailure)
		return
	}

	if cb.failureCount >= cb.failureThreshold {
		cb.fsm.Transition(EventThresholdHit)
	}
}

func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failureCount
}

func (cb *CircuitBreaker) GetFailureThreshold() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.failureThreshold
}
