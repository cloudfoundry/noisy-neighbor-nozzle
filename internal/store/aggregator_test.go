package store_test

import (
	"time"

	"code.cloudfoundry.org/noisyneighbor/internal/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aggregator", func() {
	It("pulls and stores data from a counter on a given interval", func() {
		c := store.NewCounter()

		for i := 0; i < 5; i++ {
			c.Inc("id-1")
			c.Inc("id-2")
		}

		a := store.NewAggregator(c,
			store.WithPollingInterval(50*time.Millisecond),
		)

		go a.Run()

		Eventually(a.State).Should(Equal(map[int64]map[string]uint64{
			time.Now().Unix(): map[string]uint64{
				"id-1": uint64(5),
				"id-2": uint64(5),
			},
		}))
	})
})
