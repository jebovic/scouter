package travel_test

import (
	"testing"
	"time"

	"github.com/jibei/scouter/internal/travel"
)

func TestCache_SetAndGet(t *testing.T) {
	c := travel.NewCache(time.Hour)
	c.Set("key1", "value1")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to be found")
	}
	if val != "value1" {
		t.Fatalf("expected 'value1', got %v", val)
	}
}

func TestCache_MissingKey(t *testing.T) {
	c := travel.NewCache(time.Hour)

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected key to be absent")
	}
}

func TestCache_Expiry(t *testing.T) {
	// Use a very short TTL so the entry expires immediately.
	c := travel.NewCache(1 * time.Millisecond)
	c.Set("expiring", 42)

	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("expiring")
	if ok {
		t.Fatal("expected key to have expired")
	}
}

func TestCache_OverwriteValue(t *testing.T) {
	c := travel.NewCache(time.Hour)
	c.Set("k", "first")
	c.Set("k", "second")

	val, ok := c.Get("k")
	if !ok {
		t.Fatal("expected key to be found")
	}
	if val != "second" {
		t.Fatalf("expected 'second', got %v", val)
	}
}

func TestCache_MultipleKeys(t *testing.T) {
	c := travel.NewCache(time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	for key, want := range map[string]int{"a": 1, "b": 2, "c": 3} {
		val, ok := c.Get(key)
		if !ok {
			t.Fatalf("expected key %q to be present", key)
		}
		if val.(int) != want {
			t.Fatalf("key %q: expected %d, got %v", key, want, val)
		}
	}
}

func TestCache_StoresSlice(t *testing.T) {
	c := travel.NewCache(time.Hour)
	flights := []string{"AF1234", "BA567"}
	c.Set("flights:CDG:NCE:2026-04-01", flights)

	val, ok := c.Get("flights:CDG:NCE:2026-04-01")
	if !ok {
		t.Fatal("expected flights key to be found")
	}
	got := val.([]string)
	if len(got) != 2 || got[0] != "AF1234" {
		t.Fatalf("unexpected value: %v", got)
	}
}
