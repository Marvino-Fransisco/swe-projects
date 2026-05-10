package pkg

import "sync"

type CircuitBreakerManager struct {
	mu sync.Mutex
	m  map[string]*CircuitBreaker
}

func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		m: make(map[string]*CircuitBreaker),
	}
}

func (m *CircuitBreakerManager) Get(key string) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.m[key]
}

func (m *CircuitBreakerManager) Set(key string, cb *CircuitBreaker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = cb
}

func (m *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]*CircuitBreaker, len(m.m))
	for k, v := range m.m {
		result[k] = v
	}
	return result
}
