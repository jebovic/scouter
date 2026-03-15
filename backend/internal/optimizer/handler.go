package optimizer

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/shopping"
)

// MissionBySlugGetter resolves a slug to a Mission.
type MissionBySlugGetter interface {
	GetBySlug(ctx context.Context, slug string) (*mission.Mission, error)
}

// ShoppingLister lists shopping items for a mission.
type ShoppingLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler exposes the optimizer endpoint over HTTP.
type Handler struct {
	missionGetter MissionBySlugGetter
	shoppingRepo  ShoppingLister
	agent         OptimizerAgent
	cache         *Cache
}

// NewHandler wires a Handler from the real dependencies used in production.
func NewHandler(pool *pgxpool.Pool, provider llm.Provider) *Handler {
	return &Handler{
		missionGetter: mission.NewRepository(pool),
		shoppingRepo:  shopping.NewRepository(pool),
		agent:         NewAgent(provider),
		cache:         NewCache(30 * time.Minute),
	}
}

// Optimize handles POST /api/missions/{slug}/optimize.
// Returns 422 when the mission has fewer than 2 shopping items.
func (h *Handler) Optimize(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Resolve mission.
	m, err := h.missionGetter.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Cache hit?
	if cached := h.cache.Get(m.ID.String()); cached != nil {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Load items.
	items, err := h.shoppingRepo.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if len(items) < 2 {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "Pas assez d'articles")
		return
	}

	// Run agent.
	plan, err := h.agent.Optimize(r.Context(), items)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Stamp mission ID and cache.
	plan.MissionID = m.ID.String()
	h.cache.Set(m.ID.String(), plan)

	httputil.WriteJSON(w, http.StatusOK, plan)
}
