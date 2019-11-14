package store

import (
	"container/ring"
	"errors"
	"sort"
	"sync"
	"log"
	"time"
	"os"

	"code.cloudfoundry.org/lager"
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
	data *ring.Ring

	counter         RateCounter
	pollingInterval time.Duration
	maxRateBuckets  int

	logger	lager.Logger
}

// NewAggregator will return an initialized Aggregator
func NewAggregator(c RateCounter, opts ...AggregatorOption) *Aggregator {
	a := &Aggregator{
		counter:         c,
		pollingInterval: time.Minute,
		maxRateBuckets:  10,
		logger:	         lager.NewLogger("aggregator"),
	}

	a.logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.INFO))
	for _, o := range opts {
		o(a)
	}

	a.data = ring.New(a.maxRateBuckets)
	return a
}

// Run starts the aggregator. This method will block indefinitely.
func (a *Aggregator) Run() {
	for {
		now := time.Now()
		wait := now.Add(a.pollingInterval).
			Truncate(a.pollingInterval).
			Sub(now)
		time.Sleep(wait)

		ts := time.Now().Truncate(a.pollingInterval)
		counts := a.counter.Reset()

		a.mu.Lock()
		a.data = a.data.Next()
		a.data.Value = Rate{
			Timestamp: ts.Unix(),
			Counts:    counts,
		}
		a.mu.Unlock()
	}
}

// Rates returns the current state of the aggregator.
func (a *Aggregator) Rates() Rates {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rates := make([]Rate, 0, a.data.Len())
	a.data.Next().Do(func(value interface{}) {
		if value == nil {
			return
		}

		rates = append(rates, value.(Rate))
	})

	sort.Sort(Rates(rates))

	return rates
}

// Rate returns the rates for a single time period.
func (a *Aggregator) Rate(timestamp int64) (Rate, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	err := errRateNotFound
	var rate Rate
	a.data.Next().Do(func(value interface{}) {
		if value == nil {
			a.logger.Debug("No Rate value available for bucket")
			return
		}
                a.logger.Info("Time stamp", lager.Data{"Requested TS": timestamp, "Bucket TS": value.(Rate).Timestamp})

		intervalCeil := value.(Rate).Timestamp + int64(a.pollingInterval.Seconds())
		if value.(Rate).Timestamp <= timestamp && timestamp < intervalCeil {
			rate = value.(Rate)
			err = nil
		}
	})

	return rate, err
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

// WithMaxRateBuckets returns an AggregatorOption to configure the max number
// of Rate bucketes to store.
func WithMaxRateBuckets(n int) AggregatorOption {
	return func(a *Aggregator) {
		a.maxRateBuckets = n
	}
}

// Set logger minimum log level if configured
func WithLagerLogger(s string) AggregatorOption {
	return func(a *Aggregator) {
		l, err := lager.LogLevelFromString(s)
		if err != nil {
			log.Println(err.Error() + ": default to INFO")
			l = lager.INFO
		} else {
			log.Println("Set logging level to: " + s)
		}
		a.logger.RegisterSink(lager.NewWriterSink(os.Stdout, lager.LogLevel(l)))
	}
}
