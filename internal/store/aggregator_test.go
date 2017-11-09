package store_test

import (
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/internal/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aggregator", func() {
	It("pulls and stores data from a counter on a given interval", func() {
		c := store.NewCounter()

		startTime := time.Now().Unix()
		for i := 0; i < 5; i++ {
			c.Inc("id-1")
			c.Inc("id-2")
		}

		a := store.NewAggregator(c,
			store.WithPollingInterval(time.Second),
		)

		go a.Run()

		Eventually(a.State, 3).Should(Or(
			ContainElement(store.Rate{
				Timestamp: startTime - 1,
				Counts: map[string]uint64{
					"id-1": uint64(5),
					"id-2": uint64(5),
				},
			}),
			ContainElement(store.Rate{
				Timestamp: startTime,
				Counts: map[string]uint64{
					"id-1": uint64(5),
					"id-2": uint64(5),
				},
			}),
			ContainElement(store.Rate{
				Timestamp: startTime + 1,
				Counts: map[string]uint64{
					"id-1": uint64(5),
					"id-2": uint64(5),
				},
			}),
		))
	})
})
