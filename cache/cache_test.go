package cache_test

import (
	"code.cloudfoundry.org/noisyneighbor/cache"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cache", func() {
	Describe("TopN", func() {
		It("returns a sorted list of Stats with the highest count", func() {
			c := cache.New()

			repeat(func() { c.Inc("id-1") }, 1)
			repeat(func() { c.Inc("id-2") }, 2)
			repeat(func() { c.Inc("id-3") }, 3)
			repeat(func() { c.Inc("id-4") }, 4)
			repeat(func() { c.Inc("id-5") }, 5)

			Expect(c.TopN(2)).To(Equal([]cache.Stat{
				{ID: "id-5", Count: 5},
				{ID: "id-4", Count: 4},
			}))
		})

		It("returns an empty slice with no data", func() {
			c := cache.New()

			Expect(c.TopN(10)).To(Equal([]cache.Stat{}))
		})
	})
})

func repeat(f func(), n int) {
	for i := 0; i < n; i++ {
		f()
	}
}
