package pkg

import "sync"

type BulkheadManager struct {
	mu sync.Mutex
	m  map[string]*Bulkhead
}

func NewBulkheadManager() *BulkheadManager {
	return &BulkheadManager{
		m: make(map[string]*Bulkhead),
	}
}

func (m *BulkheadManager) Get(key string) *Bulkhead {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.m[key]
}

func (m *BulkheadManager) Set(key string, b *Bulkhead) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = b
}

func (m *BulkheadManager) GetAll() map[string]*Bulkhead {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]*Bulkhead, len(m.m))
	for k, v := range m.m {
		result[k] = v
	}
	return result
}
