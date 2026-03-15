package coach

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// CachedResponse holds the response and its expiration time.
type CachedResponse struct {
	Response  CoachResponse
	ExpiresAt time.Time
}

// Cache stores responses with TTL using sync.Map.
type Cache struct {
	m       sync.Map
	ttl     time.Duration
	nowFunc func() time.Time // for testing
}

// NewCache creates a new cache with the given TTL.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		ttl:     ttl,
		nowFunc: time.Now,
	}
}

// Get retrieves a cached response if it exists and hasn't expired.
func (c *Cache) Get(missionID uuid.UUID) *CoachResponse {
	key := missionID.String()
	if raw, ok := c.m.Load(key); ok {
		cached := raw.(CachedResponse)
		if cached.ExpiresAt.After(c.nowFunc()) {
			return &cached.Response
		}
		// Expired, remove it
		c.m.Delete(key)
	}
	return nil
}

// Set stores a response in the cache.
func (c *Cache) Set(missionID uuid.UUID, resp CoachResponse) {
	key := missionID.String()
	c.m.Store(key, CachedResponse{
		Response:  resp,
		ExpiresAt: c.nowFunc().Add(c.ttl),
	})
}

// Clear removes all entries from the cache.
func (c *Cache) Clear() {
	c.m.Range(func(key, value interface{}) bool {
		c.m.Delete(key)
		return true
	})
}
