package llm

import (
	"testing"
	"time"
)

func newEntry(name string, priority int, caps Capability) ModelEntry {
	return ModelEntry{
		Name:         name,
		Provider:     &testMockProvider{},
		Capabilities: caps,
		Priority:     priority,
		Timeout:      30 * time.Second,
	}
}

func TestNewModelPool_SortsByPriorityAscending(t *testing.T) {
	pool := NewModelPool(
		newEntry("slow", 30, CapText|CapToolUse),
		newEntry("fast", 10, CapText),
		newEntry("medium", 20, CapText|CapToolUse),
	)

	if len(pool) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(pool))
	}
	if pool[0].Name != "fast" {
		t.Errorf("pool[0]: want %q, got %q", "fast", pool[0].Name)
	}
	if pool[1].Name != "medium" {
		t.Errorf("pool[1]: want %q, got %q", "medium", pool[1].Name)
	}
	if pool[2].Name != "slow" {
		t.Errorf("pool[2]: want %q, got %q", "slow", pool[2].Name)
	}
}

func TestNewModelPool_AlreadySorted(t *testing.T) {
	pool := NewModelPool(
		newEntry("a", 1, CapText),
		newEntry("b", 2, CapText),
		newEntry("c", 3, CapText),
	)
	for i, name := range []string{"a", "b", "c"} {
		if pool[i].Name != name {
			t.Errorf("pool[%d]: want %q, got %q", i, name, pool[i].Name)
		}
	}
}

func TestNewModelPool_SingleEntry(t *testing.T) {
	pool := NewModelPool(newEntry("only", 5, CapText))
	if len(pool) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(pool))
	}
	if pool[0].Name != "only" {
		t.Errorf("expected %q, got %q", "only", pool[0].Name)
	}
}

func TestNewModelPool_EmptyPool(t *testing.T) {
	pool := NewModelPool()
	if len(pool) != 0 {
		t.Errorf("expected empty pool, got %d entries", len(pool))
	}
}

func TestModelPool_ForCapabilities_ToolUseOnly(t *testing.T) {
	pool := NewModelPool(
		newEntry("text-only", 10, CapText),
		newEntry("tool-capable", 20, CapText|CapToolUse),
		newEntry("all-caps", 30, CapText|CapToolUse|CapLongCtx),
	)

	filtered := pool.ForCapabilities(CapToolUse)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 tool-capable entries, got %d", len(filtered))
	}
	if filtered[0].Name != "tool-capable" {
		t.Errorf("filtered[0]: want %q, got %q", "tool-capable", filtered[0].Name)
	}
	if filtered[1].Name != "all-caps" {
		t.Errorf("filtered[1]: want %q, got %q", "all-caps", filtered[1].Name)
	}
}

func TestModelPool_ForCapabilities_ZeroReturnsAll(t *testing.T) {
	pool := NewModelPool(
		newEntry("a", 1, CapText),
		newEntry("b", 2, CapToolUse),
	)
	all := pool.ForCapabilities(0)
	if len(all) != len(pool) {
		t.Errorf("ForCapabilities(0): want %d entries, got %d", len(pool), len(all))
	}
}

func TestModelPool_ForCapabilities_EmptyPool(t *testing.T) {
	var pool ModelPool
	filtered := pool.ForCapabilities(CapToolUse)
	if len(filtered) != 0 {
		t.Errorf("expected empty result, got %d entries", len(filtered))
	}
}

func TestModelPool_ForCapabilities_NoMatch(t *testing.T) {
	pool := NewModelPool(
		newEntry("text-only", 1, CapText),
		newEntry("text-only-2", 2, CapText),
	)
	filtered := pool.ForCapabilities(CapToolUse)
	if len(filtered) != 0 {
		t.Errorf("expected 0 entries, got %d", len(filtered))
	}
}

func TestModelPool_ForCapabilities_PreservesOrder(t *testing.T) {
	pool := NewModelPool(
		newEntry("first", 5, CapText|CapToolUse),
		newEntry("second", 10, CapText|CapToolUse),
		newEntry("third", 15, CapText|CapToolUse),
	)
	filtered := pool.ForCapabilities(CapToolUse)
	names := []string{"first", "second", "third"}
	for i, name := range names {
		if filtered[i].Name != name {
			t.Errorf("filtered[%d]: want %q, got %q", i, name, filtered[i].Name)
		}
	}
}

func TestModelPool_ForCapabilities_LongCtx(t *testing.T) {
	pool := NewModelPool(
		newEntry("short", 1, CapText|CapToolUse),
		newEntry("long", 2, CapText|CapToolUse|CapLongCtx),
	)
	filtered := pool.ForCapabilities(CapLongCtx)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(filtered))
	}
	if filtered[0].Name != "long" {
		t.Errorf("expected %q, got %q", "long", filtered[0].Name)
	}
}
