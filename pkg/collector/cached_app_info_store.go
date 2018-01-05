package collector

import (
	"log"
	"strings"
	"sync"
	"time"
)

// CachedAppInfoStore caches app info lookups against the APIStore.
type CachedAppInfoStore struct {
	store            AppInfoStore
	cacheTTL         time.Duration
	cacheLastCleared time.Time

	mu    sync.Mutex
	cache map[AppGUID]AppInfo
}

// NewCachedAppInfoStore initializes a CachedAppInfoStore.
func NewCachedAppInfoStore(s AppInfoStore, opts ...CachedAppInfoStoreOption) *CachedAppInfoStore {
	c := &CachedAppInfoStore{
		store:            s,
		cache:            make(map[AppGUID]AppInfo),
		cacheTTL:         150 * time.Second,
		cacheLastCleared: time.Now(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Lookup associates AppInfo for a particular app GUID.
func (c *CachedAppInfoStore) Lookup(guids []string) (map[AppGUID]AppInfo, error) {
	if c.cacheLastCleared.Add(c.cacheTTL).Before(time.Now()) {
		c.cache = make(map[AppGUID]AppInfo)
		c.cacheLastCleared = time.Now()
	}

	var toLookup []string
	cached := make(map[AppGUID]AppInfo)

	c.mu.Lock()
	for _, g := range guids {
		appInfo, ok := c.cache[AppGUID(g)]
		if !ok {
			toLookup = append(toLookup, g)
			continue
		}
		cached[AppGUID(g)] = appInfo
	}
	c.mu.Unlock()

	if len(toLookup) == 0 {
		return cached, nil
	}

	fresh, err := c.store.Lookup(toLookup)
	if err != nil {
		log.Printf("call to HTTP store failed: %s", err)
		return cached, nil
	}

	c.updateCache(fresh)
	return merge(fresh, cached), nil
}

func (c *CachedAppInfoStore) updateCache(fresh map[AppGUID]AppInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range fresh {
		c.cache[k] = v
	}
}

type CachedAppInfoStoreOption func(*CachedAppInfoStore)

func WithCacheTTL(cacheTTL time.Duration) CachedAppInfoStoreOption {
	return func(c *CachedAppInfoStore) {
		c.cacheTTL = cacheTTL
	}
}

// GUIDIndex is a concatentation of GUID and instance index in the format
// some-guid/some-index, e.g., 7b8228a0-cf40-42d8-a7bb-b287a88198a3/0
type GUIDIndex string

// GUID returns the GUID of the GUIDIndex
func (g GUIDIndex) GUID() string {
	return strings.Split(string(g), "/")[0]
}

// Index returns the Index of the GUIDIndex
func (g GUIDIndex) Index() string {
	parts := strings.Split(string(g), "/")
	if len(parts) < 2 {
		return "0"
	}
	return parts[1]
}

// merge combines two maps. If a key exists in both a and b, then a's value
// takes precedence.
func merge(a, b map[AppGUID]AppInfo) map[AppGUID]AppInfo {
	for k, v := range a {
		b[k] = v
	}
	return b
}
