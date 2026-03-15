package benchmark

import (
	"sync"
	"time"
)

// cacheEntry holds a BenchmarkResult with its expiry time.
type cacheEntry struct {
	value     *BenchmarkResult
	expiresAt time.Time
}

// Cache is an in-memory TTL cache for BenchmarkResult values,
// safe for concurrent use. Keyed by shopping item ID string.
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

// Get returns the cached BenchmarkResult for key and true when a
// non-expired entry exists.
func (c *Cache) Get(key string) (*BenchmarkResult, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Set stores val under key, overwriting any existing entry.
func (c *Cache) Set(key string, val *BenchmarkResult) {
	c.mu.Lock()
	c.items[key] = cacheEntry{value: val, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
