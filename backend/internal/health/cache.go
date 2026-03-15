package health

import (
	"sync"
	"time"
)

// cacheEntry holds a HealthReport and its expiry timestamp.
type cacheEntry struct {
	report    *HealthReport
	expiresAt time.Time
}

// Cache is a concurrent-safe in-memory TTL cache for HealthReport values,
// keyed by mission ID string.
type Cache struct {
	mu    sync.RWMutex
	items map[string]cacheEntry
	ttl   time.Duration
}

// NewCache returns a Cache whose entries expire after ttl.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]cacheEntry),
		ttl:   ttl,
	}
}

// Get returns the cached HealthReport for key and true when a non-expired
// entry exists.
func (c *Cache) Get(key string) (*HealthReport, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.report, true
}

// Set stores report under key, overwriting any existing entry.
func (c *Cache) Set(key string, report *HealthReport) {
	c.mu.Lock()
	c.items[key] = cacheEntry{report: report, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
