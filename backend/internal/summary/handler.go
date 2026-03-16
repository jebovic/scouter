package summary

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// MissionSlugGetter fetches a mission by slug.
type MissionSlugGetter interface {
	GetBySlug(ctx context.Context, slug string) (*mission.Mission, error)
}

// OptionLister lists options belonging to a mission by mission UUID.
type OptionLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]option.Option, error)
}

// ShoppingLister lists shopping items belonging to a mission by mission UUID.
type ShoppingLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler exposes the mission brief endpoint over HTTP.
type Handler struct {
	briefer  Briefer
	missions MissionSlugGetter
	options  OptionLister
	shopping ShoppingLister
	cache    *Cache
}

// NewHandler returns a Handler wired to the given dependencies.
func NewHandler(missions MissionSlugGetter, options OptionLister, shopping ShoppingLister, briefer Briefer, cache *Cache) *Handler {
	return &Handler{
		briefer:  briefer,
		missions: missions,
		options:  options,
		shopping: shopping,
		cache:    cache,
	}
}

// Get handles GET /api/missions/{slug}/summary.
// Returns the cached brief if available (within 1h TTL), otherwise generates
// a new one via the LLM agent, caches it, and returns it.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing mission slug")
		return
	}

	// Cache hit: return immediately without calling LLM.
	if cached, ok := h.cache.Get(slug); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Load mission.
	m, err := h.missions.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Load options and shopping items for context.
	opts, err := h.options.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	items, err := h.shopping.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Generate brief via LLM agent.
	brief, err := h.briefer.Brief(r.Context(), *m, opts, items)
	if err != nil {
		// LLM unavailable: return a degraded-but-valid DTO so the frontend
		// can render a useful state instead of an error. Do NOT cache this
		// response — the next request will retry the LLM.
		degraded := &MissionSummary{
			MissionSlug: slug,
			Bullets:     []string{"Service LLM temporairement indisponible"},
			Verdict:     "research_more",
			CachedAt:    time.Now().Unix(),
		}
		httputil.WriteJSON(w, http.StatusOK, degraded)
		return
	}

	h.cache.Set(slug, brief)
	httputil.WriteJSON(w, http.StatusOK, brief)
}
