package persona

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

const shoppingPersonaCacheTTL = 4 * time.Hour

// missionLister is the subset of mission.Repository needed by ShoppingPersonaHandler.
type missionLister interface {
	List(ctx context.Context) ([]mission.Mission, error)
}

// optionListByMission is the subset of option.Repository needed by ShoppingPersonaHandler.
type optionListByMission interface {
	ListByMission(ctx context.Context, missionID interface{ String() string }) ([]option.Option, error)
}

// shoppingListByMission is the subset of shopping.Repository needed by ShoppingPersonaHandler.
type shoppingListByMission interface {
	ListByMission(ctx context.Context, missionID interface{ String() string }) ([]shopping.Item, error)
}

// ShoppingPersonaHandler serves the GET /api/persona/shopping endpoint.
type ShoppingPersonaHandler struct {
	missionRepo mission.Repository
	optionRepo  option.Repository
	shoppingRepo shopping.Repository
	agent       *ShoppingPersonaAgent

	mu        sync.Mutex
	cached    *ShoppingPersona
	cachedAt  time.Time
}

// NewShoppingPersonaHandler creates a ShoppingPersonaHandler.
func NewShoppingPersonaHandler(
	missionRepo mission.Repository,
	optionRepo option.Repository,
	shoppingRepo shopping.Repository,
	agent *ShoppingPersonaAgent,
) *ShoppingPersonaHandler {
	return &ShoppingPersonaHandler{
		missionRepo:  missionRepo,
		optionRepo:   optionRepo,
		shoppingRepo: shoppingRepo,
		agent:        agent,
	}
}

// GetShoppingPersona handles GET /api/persona/shopping.
// Returns a ShoppingPersona built from all missions + shopping items.
// Returns a stub persona when fewer than 2 missions exist.
// Caches the result for 4 hours.
func (h *ShoppingPersonaHandler) GetShoppingPersona(w http.ResponseWriter, r *http.Request) {
	// Check cache.
	h.mu.Lock()
	if h.cached != nil && time.Since(h.cachedAt) < shoppingPersonaCacheTTL {
		p := *h.cached
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, p)
		return
	}
	h.mu.Unlock()

	ctx := r.Context()

	missions, err := h.missionRepo.List(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "shopping persona: list missions failed", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Return stub when fewer than 2 missions.
	if len(missions) < 2 {
		stub := defaultShoppingPersona
		stub.CachedAt = time.Now().Unix()
		httputil.WriteJSON(w, http.StatusOK, stub)
		return
	}

	summary := h.buildSummary(ctx, missions)

	persona, err := h.agent.Analyze(ctx, summary)
	if err != nil {
		slog.ErrorContext(ctx, "shopping persona: agent analyze failed", "err", err)
		stub := defaultShoppingPersona
		stub.CachedAt = time.Now().Unix()
		httputil.WriteJSON(w, http.StatusOK, stub)
		return
	}

	persona.CachedAt = time.Now().Unix()

	h.mu.Lock()
	h.cached = &persona
	h.cachedAt = time.Now()
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, persona)
}

// buildSummary aggregates missions and shopping item data into a shoppingSummary.
func (h *ShoppingPersonaHandler) buildSummary(ctx context.Context, missions []mission.Mission) shoppingSummary {
	s := shoppingSummary{
		missionCount:  len(missions),
		categories:   make(map[string]int),
		missionPhases: make(map[string]int),
	}

	for _, m := range missions {
		if m.Category != "" {
			s.categories[m.Category]++
		}
		if m.Phase != "" {
			s.missionPhases[m.Phase]++
		}

		items, err := h.shoppingRepo.ListByMission(ctx, m.ID)
		if err != nil {
			continue
		}
		for _, item := range items {
			s.totalSpent += item.Price
		}
	}

	return s
}
