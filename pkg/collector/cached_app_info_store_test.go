package collector_test

import (
	"errors"
	"time"

	"code.cloudfoundry.org/noisy-neighbor-nozzle/pkg/collector"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CachedAppInfoStore", func() {
	It("finds app info for a GUID by quering the API store", func() {
		expected := map[collector.AppGUID]collector.AppInfo{
			"app-guid-1": collector.AppInfo{
				Name:  "some-name",
				Space: "some-space",
				Org:   "some-org",
			},
		}
		apiStore := &spyAPIStore{lookupReturns: expected}
		store := collector.NewCachedAppInfoStore(apiStore)

		actual, _ := store.Lookup([]string{"app-guid-1"})

		Expect(apiStore.lookupCalled).To(Equal(1))
		Expect(apiStore.lookupGUIDInstances).To(Equal([]string{"app-guid-1"}))
		Expect(actual).To(Equal(expected))
	})

	It("caches app info to prevent unnecessary API calls", func() {
		expected := map[collector.AppGUID]collector.AppInfo{
			"app-guid-1": collector.AppInfo{
				Name:  "some-name",
				Space: "some-space",
				Org:   "some-org",
			},
		}
		apiStore := &spyAPIStore{lookupReturns: expected}
		store := collector.NewCachedAppInfoStore(apiStore)

		first, _ := store.Lookup([]string{"app-guid-1"})
		second, _ := store.Lookup([]string{"app-guid-1"})

		Expect(apiStore.lookupCalled).To(Equal(1))
		Expect(first).To(Equal(expected))
		Expect(second).To(Equal(expected))
	})

	It("eventually updates the cache", func() {
		expected := map[collector.AppGUID]collector.AppInfo{
			"app-guid-1": collector.AppInfo{
				Name:  "some-name",
				Space: "some-space",
				Org:   "some-org",
			},
		}

		apiStore := &spyAPIStore{lookupReturns: expected}
		store := collector.NewCachedAppInfoStore(apiStore, collector.WithCacheTTL(100*time.Millisecond))

		Eventually(func() int {
			_, _ = store.Lookup([]string{"app-guid-1"})
			return apiStore.lookupCalled
		}).Should(BeNumerically(">", 1))
	})

	It("returns the cache when the API store fails", func() {
		apiStore := &spyAPIStore{lookupError: errors.New("HTTP request failed")}
		store := collector.NewCachedAppInfoStore(apiStore)

		actual, err := store.Lookup([]string{"app-guid-1"})

		Expect(err).NotTo(HaveOccurred())
		emptyCache := make(map[collector.AppGUID]collector.AppInfo)
		Expect(actual).To(Equal(emptyCache))
	})

	// NOTE This test assumes test invocations occur with `-race`.
	It("supports thread safe read access", func() {
		apiStore := &spyAPIStore{
			lookupReturns: make(map[collector.AppGUID]collector.AppInfo),
		}
		store := collector.NewCachedAppInfoStore(apiStore)

		go func() {
			for {
				store.Lookup([]string{"app-guid-1"})
			}
		}()

		store.Lookup([]string{"app-guid-1"})
	})

	Describe("GUIDIndex", func() {
		It("returns a GUID", func() {
			g := collector.GUIDIndex("12abc/0")

			id := g.GUID()

			Expect(id).To(Equal("12abc"))
		})

		It("returns the entire value when no slash is present", func() {
			g := collector.GUIDIndex("12abc")

			guid := g.GUID()

			Expect(guid).To(Equal("12abc"))
		})

		It("returns an index", func() {
			g := collector.GUIDIndex("12abc/0")

			id := g.Index()

			Expect(id).To(Equal("0"))
		})

		It("returns 0 when no index is found", func() {
			g := collector.GUIDIndex("12abc")

			id := g.Index()

			Expect(id).To(Equal("0"))
		})
	})
})

type spyAPIStore struct {
	lookupCalled        int
	lookupGUIDInstances []string
	lookupReturns       map[collector.AppGUID]collector.AppInfo
	lookupError         error
}

func (s *spyAPIStore) Lookup(guids []string) (map[collector.AppGUID]collector.AppInfo, error) {
	s.lookupCalled++
	s.lookupGUIDInstances = guids
	return s.lookupReturns, s.lookupError
}
