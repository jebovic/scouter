package llm

import (
	"sort"
	"time"
)

// ModelEntry describes a single model in the pool.
type ModelEntry struct {
	// Name is a human-readable identifier (e.g. "qwen3:14b").
	Name string
	// Provider is the transport that handles this model.
	Provider Provider
	// Capabilities lists what this model supports.
	Capabilities Capability
	// Priority controls selection order: lower value = tried first.
	Priority int
	// Timeout is the per-request deadline for this model.
	Timeout time.Duration
	// SystemPromptPrefix, when non-empty, is prepended as a system message before
	// every request routed to this model. Use "/no_think" for Qwen3 models to
	// disable chain-of-thought reasoning in agent loops.
	SystemPromptPrefix string
}

// ModelPool is an ordered slice of ModelEntry values, sorted by Priority ascending.
type ModelPool []ModelEntry

// NewModelPool creates a ModelPool from the given entries, sorting by Priority ascending.
func NewModelPool(entries ...ModelEntry) ModelPool {
	pool := make(ModelPool, len(entries))
	copy(pool, entries)
	sort.Slice(pool, func(i, j int) bool {
		return pool[i].Priority < pool[j].Priority
	})
	return pool
}

// ForCapabilities returns a new ModelPool containing only entries that satisfy
// all bits in required. Order is preserved.
// If required is zero, all entries are returned.
func (p ModelPool) ForCapabilities(required Capability) ModelPool {
	if required == 0 {
		out := make(ModelPool, len(p))
		copy(out, p)
		return out
	}
	out := make(ModelPool, 0, len(p))
	for _, e := range p {
		if e.Capabilities.Has(required) {
			out = append(out, e)
		}
	}
	return out
}
