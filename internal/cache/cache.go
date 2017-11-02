package cache

import (
	"sort"
	"sync"
)

// Cache stores data about the number of logs emitted per application.
type Cache struct {
	mu   sync.RWMutex
	data map[string]int64
}

// New returns an initialized cache.
func New() *Cache {
	return &Cache{
		data: make(map[string]int64),
	}
}

// Inc increments the value for a given ID.
func (c *Cache) Inc(id string) {
	c.mu.Lock()
	c.data[id]++
	c.mu.Unlock()
}

// TopN returns a sorted list of the IDs with the highest count.
func (c *Cache) TopN(n int) []Stat {
	c.mu.RLock()
	stats := make([]Stat, 0, len(c.data))
	for k, v := range c.data {
		stats = append(stats, Stat{k, v})
	}
	c.mu.RUnlock()

	sort.Sort(Stats(stats))

	if len(stats) < n {
		return stats
	}

	return stats[0:n]
}
