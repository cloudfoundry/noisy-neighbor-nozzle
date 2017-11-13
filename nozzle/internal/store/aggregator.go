package store

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	errRateNotFound = errors.New("rate not found")
)

// RateCounter is the interface the Aggregator will poll data from.
type RateCounter interface {
	Reset() map[string]uint64
}

// Aggregator will pull from the Counter on a given interval and store rates for
// that interval.
type Aggregator struct {
	mu   sync.RWMutex
	data map[int64]map[string]uint64

	counter         RateCounter
	pollingInterval time.Duration
}

// NewAggregator will return an initialized Aggregator
func NewAggregator(c RateCounter, opts ...AggregatorOption) *Aggregator {
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

// Rates returns the current state of the aggregator.
func (a *Aggregator) Rates() Rates {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rates := make([]Rate, 0, len(a.data))
	for ts, data := range a.data {
		r := Rate{
			Timestamp: ts,
			Counts:    make(map[string]uint64),
		}

		for id, count := range data {
			r.Counts[id] = count
		}

		rates = append(rates, r)
	}

	sort.Sort(Rates(rates))

	return rates
}

// Rate returns the rates for a single time period.
func (a *Aggregator) Rate(timestamp int64) (Rate, error) {
	counts, ok := a.data[timestamp]
	if !ok {
		return Rate{}, errRateNotFound
	}

	return Rate{
		Timestamp: timestamp,
		Counts:    counts,
	}, nil
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
