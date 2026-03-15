package rebalancer

import (
	"sync"
	"time"
)

// cacheEntry holds a plan and its expiry timestamp.
type cacheEntry struct {
	plan      *RebalancePlan
	expiresAt time.Time
}

// Cache is a thread-safe TTL store for RebalancePlan values keyed by a
// canonical string (typically a hash of sorted envelope IDs).
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
	nowFunc func() time.Time
}

// NewCache returns a Cache with the given time-to-live.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		nowFunc: time.Now,
	}
}

// Get returns the cached plan for key, or nil if absent or expired.
func (c *Cache) Get(key string) *RebalancePlan {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil
	}
	if e.expiresAt.Before(c.nowFunc()) {
		c.mu.Lock()
		delete(c.entries, key)
		c.mu.Unlock()
		return nil
	}
	return e.plan
}

// Set stores plan under key for c.ttl.
func (c *Cache) Set(key string, plan *RebalancePlan) {
	c.mu.Lock()
	c.entries[key] = cacheEntry{
		plan:      plan,
		expiresAt: c.nowFunc().Add(c.ttl),
	}
	c.mu.Unlock()
}
