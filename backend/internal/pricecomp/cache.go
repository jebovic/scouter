package pricecomp

import (
	"sync"
	"time"
)

// cacheEntry holds a cached PriceComparison and its expiry time.
type cacheEntry struct {
	value     *PriceComparison
	expiresAt time.Time
}

// Cache is a simple in-memory TTL cache for PriceComparison values, safe for
// concurrent use. Keyed by option ID string.
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

// Get returns the cached PriceComparison for key and true when a non-expired
// entry exists.
func (c *Cache) Get(key string) (*PriceComparison, bool) {
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
func (c *Cache) Set(key string, val *PriceComparison) {
	c.mu.Lock()
	c.items[key] = cacheEntry{value: val, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

// Delete removes the entry for key. It is a no-op if the key is not present.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}
