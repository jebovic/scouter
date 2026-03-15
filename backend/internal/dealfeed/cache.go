package dealfeed

import (
	"sync"
	"time"
)

// cacheEntry holds a cached PromoFeedResult with its expiry.
type cacheEntry struct {
	result    *PromoFeedResult
	expiresAt time.Time
}

// Cache is a thread-safe TTL cache for promo feed results, keyed by query+category string.
type Cache struct {
	m       sync.Map
	ttl     time.Duration
	nowFunc func() time.Time
}

// NewCache returns a new Cache with the given TTL using the real clock.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, nowFunc: time.Now}
}

// NewCacheWithNow returns a new Cache with a custom clock for testing.
func NewCacheWithNow(ttl time.Duration, nowFunc func() time.Time) *Cache {
	return &Cache{ttl: ttl, nowFunc: nowFunc}
}

// Get retrieves a cached result for the given key. Returns nil if absent or expired.
func (c *Cache) Get(key string) *PromoFeedResult {
	raw, ok := c.m.Load(key)
	if !ok {
		return nil
	}
	e := raw.(cacheEntry)
	if e.expiresAt.After(c.nowFunc()) {
		return e.result
	}
	c.m.Delete(key)
	return nil
}

// Set stores result in the cache under key for c.ttl.
func (c *Cache) Set(key string, result *PromoFeedResult) {
	c.m.Store(key, cacheEntry{
		result:    result,
		expiresAt: c.nowFunc().Add(c.ttl),
	})
}

// cacheKey builds a deterministic cache key from query and category.
func cacheKey(query, category string) string {
	return query + "|" + category
}
