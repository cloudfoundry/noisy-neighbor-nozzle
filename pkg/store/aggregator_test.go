package store_test

import (
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Aggregator", func() {
	Describe("Rates", func() {
		It("pulls and stores data from a counter on a given interval", func() {
			a := store.NewAggregator(stubRateCounter{},
				store.WithPollingInterval(time.Second),
			)

			go a.Run()

			var rates []store.Rate
			f := func() []store.Rate {
				rates = a.Rates()
				return rates
			}
			Eventually(f).ShouldNot(BeEmpty())

			Expect(rates).To(HaveLen(1))
			Expect(rates[0].Timestamp).To(BeNumerically("~", time.Now().Unix(), 1))
			Expect(rates[0].Counts).To(Equal(map[string]uint64{
				"id-1": uint64(5),
				"id-2": uint64(5),
			}))
		})

		It("prunes older rates", func() {
			a := store.NewAggregator(stubRateCounter{},
				store.WithPollingInterval(1*time.Millisecond),
				store.WithMaxRateBuckets(2),
			)

			go a.Run()

			Eventually(a.Rates).Should(HaveLen(2))
			Consistently(a.Rates).Should(HaveLen(2))
		})
	})

	Describe("Rate", func() {
		It("returns a the rates for a single timestamp", func() {
			a := store.NewAggregator(stubRateCounter{},
				store.WithPollingInterval(20*time.Millisecond),
			)

			go a.Run()

			ts := time.Now().Truncate(20 * time.Millisecond).Unix()
			var rate store.Rate
			Eventually(func() error {
				var err error
				rate, err = a.Rate(ts)

				return err
			}).Should(Succeed())

			Expect(rate).To(Equal(store.Rate{
				Timestamp: ts,
				Counts: map[string]uint64{
					"id-1": uint64(5),
					"id-2": uint64(5),
				},
			}))
		})

		It("returns an error if no rate is found for the timestamp", func() {
			a := store.NewAggregator(stubRateCounter{},
				store.WithPollingInterval(20*time.Millisecond),
			)

			_, err := a.Rate(time.Now().Unix())
			Expect(err).To(MatchError("rate not found"))
		})
	})
})

type stubRateCounter struct{}

func (s stubRateCounter) Reset() map[string]uint64 {
	return map[string]uint64{
		"id-1": uint64(5),
		"id-2": uint64(5),
	}
}
