package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBulkhead_AllowsUpToMax(t *testing.T) {
	bh := NewBulkhead(2, 100*time.Millisecond)
	assert.NoError(t, bh.Acquire()) // slot 1
	assert.NoError(t, bh.Acquire()) // slot 2
	assert.Error(t, bh.Acquire())   // slot 3 → rejected after timeout
	bh.Release()                    // free slot 1
	assert.NoError(t, bh.Acquire()) // slot 1 again
}
func TestBulkhead_ActiveConnections(t *testing.T) {
	bh := NewBulkhead(5, 50*time.Millisecond)
	bh.Acquire()
	bh.Acquire()
	assert.Equal(t, 2, bh.ActiveConnections())
}
