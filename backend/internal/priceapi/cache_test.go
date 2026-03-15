package priceapi

import (
	"sync"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	c := NewCache(5 * time.Minute)
	listings := []PriceListing{
		{Source: "test", ProductName: "Widget", Price: 9.99, Currency: "EUR"},
	}
	key := c.Key("testprovider", "Widget", "food")
	c.Set(key, listings)

	got, ok := c.Get(key)
	if !ok {
		t.Fatal("expected cache hit, got miss")
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(got))
	}
	if got[0].ProductName != "Widget" {
		t.Errorf("expected ProductName=Widget, got %q", got[0].ProductName)
	}
}

func TestCache_GetMiss(t *testing.T) {
	c := NewCache(5 * time.Minute)
	_, ok := c.Get("nonexistent-key")
	if ok {
		t.Fatal("expected cache miss, got hit")
	}
}

func TestCache_ExpiredEntry(t *testing.T) {
	c := NewCache(1 * time.Millisecond)
	key := c.Key("prov", "Item", "grocery")
	c.Set(key, []PriceListing{{Source: "prov", ProductName: "Item"}})

	// Wait for expiry.
	time.Sleep(10 * time.Millisecond)

	_, ok := c.Get(key)
	if ok {
		t.Fatal("expected expired entry to return miss")
	}
}

func TestCache_EvictionOnSet(t *testing.T) {
	// Short TTL so entries expire quickly.
	c := NewCache(1 * time.Millisecond)
	key1 := c.Key("prov", "Item1", "food")
	key2 := c.Key("prov", "Item2", "food")

	c.Set(key1, []PriceListing{{Source: "prov", ProductName: "Item1"}})

	// Let key1 expire.
	time.Sleep(10 * time.Millisecond)

	// Set key2 — this triggers eviction of expired entries.
	c.Set(key2, []PriceListing{{Source: "prov", ProductName: "Item2"}})

	c.mu.RLock()
	_, key1Present := c.entries[key1]
	c.mu.RUnlock()

	if key1Present {
		t.Error("expected expired key1 to be evicted after Set")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := NewCache(5 * time.Minute)
	var wg sync.WaitGroup
	const goroutines = 50

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := c.Key("prov", "Item", "food")
			c.Set(key, []PriceListing{{Source: "prov", Price: float64(n)}})
			c.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestCache_ImmutableReturn(t *testing.T) {
	c := NewCache(5 * time.Minute)
	original := []PriceListing{
		{Source: "prov", ProductName: "Widget", Price: 5.00},
	}
	key := c.Key("prov", "Widget", "food")
	c.Set(key, original)

	got, ok := c.Get(key)
	if !ok {
		t.Fatal("expected cache hit")
	}

	// Mutate the returned slice — cache should be unaffected.
	got[0].Price = 999.99
	got[0].ProductName = "Mutated"

	got2, ok := c.Get(key)
	if !ok {
		t.Fatal("expected second cache hit")
	}
	if got2[0].Price != 5.00 {
		t.Errorf("mutation affected cache: got price %f, want 5.00", got2[0].Price)
	}
	if got2[0].ProductName != "Widget" {
		t.Errorf("mutation affected cache: got name %q, want Widget", got2[0].ProductName)
	}
}

func TestCache_Key(t *testing.T) {
	c := NewCache(time.Minute)
	tests := []struct {
		provider    string
		productName string
		category    string
		wantSuffix  string
	}{
		{"Provider", "Widget Pro", "Food", "provider:widget pro:food"},
		{"off", "café au lait", "grocery", "off:café au lait:grocery"},
	}
	for _, tt := range tests {
		got := c.Key(tt.provider, tt.productName, tt.category)
		if got != tt.wantSuffix {
			t.Errorf("Key(%q,%q,%q) = %q, want %q", tt.provider, tt.productName, tt.category, got, tt.wantSuffix)
		}
	}
}
