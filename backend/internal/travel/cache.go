package travel

import (
	"sync"
	"time"
)

// entry holds a cached value and the time at which it expires.
type entry struct {
	value     any
	expiresAt time.Time
}

// Cache is a simple in-memory TTL cache safe for concurrent use.
type Cache struct {
	mu    sync.RWMutex
	items map[string]entry
	ttl   time.Duration
}

// NewCache returns a Cache whose entries expire after ttl.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		items: make(map[string]entry),
		ttl:   ttl,
	}
}

// Get returns the cached value for key and true when a non-expired entry exists.
func (c *Cache) Get(key string) (any, bool) {
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
func (c *Cache) Set(key string, val any) {
	c.mu.Lock()
	c.items[key] = entry{value: val, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
