package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger is an optional interface providers can implement for liveness checks.
type Pinger interface {
	Ping(ctx context.Context) error
}

// ModelStatus holds the health status of a single model entry.
type ModelStatus struct {
	Name         string   `json:"name"`
	Healthy      bool     `json:"healthy"`
	CircuitState string   `json:"circuit_state"` // "open" | "half_open" | "closed"
	Capabilities []string `json:"capabilities"`
}

// PoolHealth is the aggregate health report for the entire model pool.
type PoolHealth struct {
	Models    []ModelStatus `json:"models"`
	CheckedAt time.Time     `json:"checked_at"`
}

// HealthChecker evaluates the liveness of every model in a pool.
type HealthChecker struct {
	pool     ModelPool
	breakers map[string]*circuitBreaker
}

// NewHealthChecker creates a HealthChecker. breakers may be nil (all circuits assumed closed).
func NewHealthChecker(pool ModelPool, breakers map[string]*circuitBreaker) *HealthChecker {
	return &HealthChecker{pool: pool, breakers: breakers}
}

// Check probes every model and returns a PoolHealth snapshot.
func (h *HealthChecker) Check(ctx context.Context) PoolHealth {
	statuses := make([]ModelStatus, 0, len(h.pool))

	for _, entry := range h.pool {
		circuitState := "closed"
		if h.breakers != nil {
			if cb, ok := h.breakers[entry.Name]; ok {
				circuitState = cb.state()
			}
		}
		circuitOpen := circuitState == "open"

		healthy := false
		if pinger, ok := entry.Provider.(Pinger); ok {
			pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
			err := pinger.Ping(pingCtx)
			pingCancel()
			healthy = err == nil && !circuitOpen
		} else {
			healthy = !circuitOpen
		}

		statuses = append(statuses, ModelStatus{
			Name:         entry.Name,
			Healthy:      healthy,
			CircuitState: circuitState,
			Capabilities: capabilityNames(entry.Capabilities),
		})
	}

	return PoolHealth{
		Models:    statuses,
		CheckedAt: time.Now().UTC(),
	}
}

// capabilityNames converts a Capability bitmask to an ordered slice of names.
func capabilityNames(cap Capability) []string {
	type named struct {
		bit  Capability
		name string
	}
	all := []named{
		{CapText, "text"},
		{CapToolUse, "tool_use"},
		{CapLongCtx, "long_ctx"},
	}
	var names []string
	for _, n := range all {
		if cap.Has(n.bit) {
			names = append(names, n.name)
		}
	}
	return names
}

// HealthHandler returns an http.HandlerFunc that writes a PoolHealth JSON response.
func HealthHandler(checker *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := checker.Check(r.Context())
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(health)
	}
}
