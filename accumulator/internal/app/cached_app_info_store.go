package app

import (
	"log"
	"strings"
	"sync"
)

// CachedAppInfoStore caches app info lookups against the APIStore.
type CachedAppInfoStore struct {
	store AppInfoStore

	mu    sync.Mutex
	cache map[AppGUID]AppInfo
}

// NewCachedAppInfoStore initializes a CachedAppInfoStore.
func NewCachedAppInfoStore(s AppInfoStore) *CachedAppInfoStore {
	return &CachedAppInfoStore{
		store: s,
		cache: make(map[AppGUID]AppInfo),
	}
}

// GUIDIndex is a concatentation of GUID and instance index in the format
// some-guid/some-index, e.g., 7b8228a0-cf40-42d8-a7bb-b287a88198a3/0
type GUIDIndex string

func (g GUIDIndex) GUID() string {
	return strings.Split(string(g), "/")[0]
}

func (g GUIDIndex) Index() string {
	parts := strings.Split(string(g), "/")
	if len(parts) < 2 {
		return "0"
	}
	return parts[1]
}

// Lookup associates AppInfo for a particular app GUID.
func (c *CachedAppInfoStore) Lookup(guids []string) (map[AppGUID]AppInfo, error) {
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

// merge combines two maps. If a key exists in both a and b, then a's value
// takes precedence.
func merge(a, b map[AppGUID]AppInfo) map[AppGUID]AppInfo {
	for k, v := range a {
		b[k] = v
	}
	return b
}
