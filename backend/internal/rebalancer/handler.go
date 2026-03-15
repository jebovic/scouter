package rebalancer

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

// envelopeRepo is the subset of envelope.Repository used by the handler.
type envelopeRepo interface {
	FindAll(ctx context.Context) ([]envelope.Envelope, error)
}

// missionLister is the subset of mission.Repository used by the handler.
type missionLister interface {
	List(ctx context.Context) ([]mission.Mission, error)
}

// agentRunner is the interface satisfied by *Agent.
type agentRunner interface {
	Rebalance(ctx context.Context, envelopes []envelope.Envelope, missions []mission.Mission) (*RebalancePlan, error)
}

// Handler exposes the budget rebalancer endpoint.
type Handler struct {
	envelopeRepo envelopeRepo
	missionRepo  missionLister
	agent        agentRunner
	cache        *Cache
}

// NewHandler constructs a Handler with the given dependencies.
func NewHandler(envelopeRepo envelopeRepo, missionRepo missionLister, agent agentRunner) *Handler {
	return &Handler{
		envelopeRepo: envelopeRepo,
		missionRepo:  missionRepo,
		agent:        agent,
		cache:        NewCache(6 * time.Hour),
	}
}

// Get handles GET /api/budget/rebalance.
// It gathers all envelopes + missions, calls the rebalancer agent (with a 6 h
// cache keyed by the sorted set of envelope IDs), and returns a RebalancePlan.
// If no envelopes exist, it returns 200 with an empty adjustments list.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	envelopes, err := h.envelopeRepo.FindAll(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "rebalancer: fetch envelopes", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Fast path: no envelopes — return empty plan immediately.
	if len(envelopes) == 0 {
		httputil.WriteJSON(w, http.StatusOK, &RebalancePlan{
			Adjustments:    []EnvelopeAdjustment{},
			Summary:        "",
			TotalCurrent:   0,
			TotalSuggested: 0,
			CachedAt:       0,
		})
		return
	}

	cacheKey := envelopeCacheKey(envelopes)
	if cached := h.cache.Get(cacheKey); cached != nil {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	missions, err := h.missionRepo.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "rebalancer: fetch missions", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	plan, err := h.agent.Rebalance(ctx, envelopes, missions)
	if err != nil {
		slog.ErrorContext(ctx, "rebalancer: agent call failed", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Set(cacheKey, plan)
	httputil.WriteJSON(w, http.StatusOK, plan)
}

// envelopeCacheKey returns a stable SHA-256 hex digest of the sorted envelope
// IDs. This ensures the cache is invalidated when envelopes are added or removed.
func envelopeCacheKey(envelopes []envelope.Envelope) string {
	ids := make([]string, len(envelopes))
	for i, env := range envelopes {
		ids[i] = env.ID.String()
	}
	sort.Strings(ids)
	h := sha256.New()
	for _, id := range ids {
		fmt.Fprint(h, id)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
