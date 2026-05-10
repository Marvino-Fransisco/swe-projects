package pkg

import (
	"fmt"
	"time"
)

type Bulkhead struct {
	sem          chan struct{}
	maxConns     int
	queueTimeout time.Duration
}

func NewBulkhead(maxConnections int, queueTimeout time.Duration) *Bulkhead {
	if maxConnections <= 0 {
		maxConnections = 1
	}
	return &Bulkhead{
		sem:          make(chan struct{}, maxConnections),
		maxConns:     maxConnections,
		queueTimeout: queueTimeout,
	}
}

func (b *Bulkhead) Acquire() error {
	select {
	case b.sem <- struct{}{}:
		return nil
	default:
	}

	select {
	case b.sem <- struct{}{}:
		return nil
	case <-time.After(b.queueTimeout):
		return fmt.Errorf("bulkhead full: max %d concurrent connections reached after waiting %v", b.maxConns, b.queueTimeout)
	}
}

func (b *Bulkhead) Release() {
	<-b.sem
}

func (b *Bulkhead) MaxConnections() int {
	return b.maxConns
}

func (b *Bulkhead) ActiveConnections() int {
	return len(b.sem)
}
