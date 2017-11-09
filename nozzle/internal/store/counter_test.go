package store_test

import (
	"code.cloudfoundry.org/noisy-neighbor-nozzle/nozzle/internal/store"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Counter", func() {
	Describe("Reset", func() {
		It("returns all of the current counts", func() {
			c := store.NewCounter()

			repeat(func() { c.Inc("id-1") }, 1)
			repeat(func() { c.Inc("id-2") }, 2)
			repeat(func() { c.Inc("id-3") }, 3)
			repeat(func() { c.Inc("id-4") }, 4)
			repeat(func() { c.Inc("id-5") }, 5)

			counts := c.Reset()
			Expect(counts).To(HaveKeyWithValue("id-1", uint64(1)))
			Expect(counts).To(HaveKeyWithValue("id-2", uint64(2)))
			Expect(counts).To(HaveKeyWithValue("id-3", uint64(3)))
			Expect(counts).To(HaveKeyWithValue("id-4", uint64(4)))
			Expect(counts).To(HaveKeyWithValue("id-5", uint64(5)))
		})

		It("resets the current counts", func() {
			c := store.NewCounter()

			repeat(func() { c.Inc("id-1") }, 1)

			_ = c.Reset()
			Expect(c.Reset()).To(BeEmpty())
		})
	})
})

func repeat(f func(), n int) {
	for i := 0; i < n; i++ {
		f()
	}
}
