package pricealertrule

import (
	"fmt"
	"sync"
)

// Repository is an in-memory store for alert rules, keyed by rule ID.
type Repository struct {
	mu    sync.RWMutex
	rules map[string]AlertRule // key: rule ID
}

// NewRepository returns an initialised in-memory repository.
func NewRepository() *Repository {
	return &Repository{
		rules: make(map[string]AlertRule),
	}
}

// AddRule persists a new rule. Returns an error if a rule with the same ID
// already exists (should not happen with UUID generation).
func (r *Repository) AddRule(rule AlertRule) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rules[rule.ID]; exists {
		return fmt.Errorf("rule %s already exists", rule.ID)
	}
	r.rules[rule.ID] = rule
	return nil
}

// ListByItem returns all rules for a given item ID, sorted by creation order
// (insertion order is not guaranteed in a map, so callers should not rely on
// deterministic ordering).
func (r *Repository) ListByItem(itemID string) []AlertRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]AlertRule, 0)
	for _, rule := range r.rules {
		if rule.ItemID == itemID {
			out = append(out, rule)
		}
	}
	return out
}

// DeleteRule removes a rule by ID. Returns an error if the rule does not exist.
func (r *Repository) DeleteRule(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.rules[id]; !exists {
		return fmt.Errorf("rule %s not found", id)
	}
	delete(r.rules, id)
	return nil
}

// MarkTriggered sets Triggered=true for the rule with the given ID.
func (r *Repository) MarkTriggered(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	rule, exists := r.rules[id]
	if !exists {
		return fmt.Errorf("rule %s not found", id)
	}
	rule.Triggered = true
	r.rules[id] = rule
	return nil
}

// ListAll returns every rule in the store.
func (r *Repository) ListAll() []AlertRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]AlertRule, 0, len(r.rules))
	for _, rule := range r.rules {
		out = append(out, rule)
	}
	return out
}
