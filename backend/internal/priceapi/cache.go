package priceapi

import (
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	listings  []PriceListing
	expiresAt time.Time
}

// Cache is a thread-safe in-memory TTL cache for price listings.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

// NewCache creates a new Cache with the given TTL duration.
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

// Get returns a copy of the cached listings for key.
// Returns false if the key is missing or the entry has expired.
func (c *Cache) Get(key string) ([]PriceListing, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return nil, false
	}
	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	// Return a defensive copy so callers cannot mutate the cached slice.
	out := make([]PriceListing, len(entry.listings))
	copy(out, entry.listings)
	return out, true
}

// Set stores a copy of listings under key and evicts all expired entries.
func (c *Cache) Set(key string, listings []PriceListing) {
	// Defensive copy to prevent the caller's slice from aliasing the cache.
	stored := make([]PriceListing, len(listings))
	copy(stored, listings)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict expired entries on every write to bound memory use.
	now := time.Now()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}

	c.entries[key] = cacheEntry{
		listings:  stored,
		expiresAt: now.Add(c.ttl),
	}
}

// Key returns a lowercase normalized cache key from the provider name, product name,
// and category.
func (c *Cache) Key(provider, productName, category string) string {
	return strings.ToLower(provider) + ":" + strings.ToLower(productName) + ":" + strings.ToLower(category)
}
