package store

import (
	"sync"
	"time"
)

// Aggregator will pull from the Counter on a given interval and store rates for
// that interval.
type Aggregator struct {
	mu   sync.RWMutex
	data map[int64]map[string]uint64

	counter         *Counter
	pollingInterval time.Duration
}

// NewAggregator will return an initialized Aggregator
func NewAggregator(c *Counter, opts ...AggregatorOption) *Aggregator {
	a := &Aggregator{
		data:            make(map[int64]map[string]uint64),
		counter:         c,
		pollingInterval: time.Minute,
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

// Run starts the aggregator. This method will block indefinately.
func (a *Aggregator) Run() {
	for {
		now := time.Now()
		wait := now.Add(a.pollingInterval).
			Truncate(a.pollingInterval).
			Sub(now)
		time.Sleep(wait)

		ts := time.Now().Truncate(a.pollingInterval)
		data := a.counter.Reset()

		a.mu.Lock()
		a.data[ts.Unix()] = data
		a.mu.Unlock()

		// TODO: Remove old keys
	}
}

// State returns the current state of the aggregator.
func (a *Aggregator) State() map[int64]map[string]uint64 {
	a.mu.RLock()
	defer a.mu.RUnlock()

	state := make(map[int64]map[string]uint64)
	for ts, data := range a.data {
		state[ts] = make(map[string]uint64)
		for id, count := range data {
			state[ts][id] = count
		}
	}

	return state
}

// AggregatorOption are funcs that can be used to configure an Aggregator at
// initialization.
type AggregatorOption func(a *Aggregator)

// WithPollingInterval returns an AggregatorOption to configure the polling
// interval. The polling interval determines how often the aggregator will poll
// data from the counter. The polling interval is also isued to determine the
// amount of time that is used for the rate. E.g. If the polling interval is
// 1 minute, then rates are calculated as number of logs per minute.
func WithPollingInterval(d time.Duration) AggregatorOption {
	return func(a *Aggregator) {
		a.pollingInterval = d
	}
}
