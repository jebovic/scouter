package optimizer

import (
	"sync"
	"time"
)

// entry holds a cached plan with its expiry.
type entry struct {
	plan      *OptimizedPlan
	expiresAt time.Time
}

// Cache is a thread-safe TTL cache for optimized plans, keyed by mission ID string.
type Cache struct {
	m       sync.Map
	ttl     time.Duration
	nowFunc func() time.Time
}

// NewCache returns a new Cache with the given TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl, nowFunc: time.Now}
}

// Get retrieves a cached plan for missionID. Returns nil if absent or expired.
func (c *Cache) Get(missionID string) *OptimizedPlan {
	if raw, ok := c.m.Load(missionID); ok {
		e := raw.(entry)
		if e.expiresAt.After(c.nowFunc()) {
			return e.plan
		}
		c.m.Delete(missionID)
	}
	return nil
}

// Set stores plan in the cache under missionID for c.ttl.
func (c *Cache) Set(missionID string, plan *OptimizedPlan) {
	c.m.Store(missionID, entry{
		plan:      plan,
		expiresAt: c.nowFunc().Add(c.ttl),
	})
}
