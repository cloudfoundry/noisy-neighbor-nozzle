package store

import (
	"sync"
)

// Counter stores data about the number of logs emitted per application.
type Counter struct {
	mu   sync.RWMutex
	data map[string]uint64
}

// NewCounter returns an initialized counter.
func NewCounter() *Counter {
	return &Counter{
		data: make(map[string]uint64),
	}
}

// Inc increments the value for a given ID.
func (c *Counter) Inc(id string) {
	c.mu.Lock()
	c.data[id]++
	c.mu.Unlock()
}

// Reset returns the current counts while replacing the current counts with an
// empty map.
func (c *Counter) Reset() map[string]uint64 {
	c.mu.Lock()
	d := c.data
	c.data = make(map[string]uint64)
	c.mu.Unlock()

	return d
}
