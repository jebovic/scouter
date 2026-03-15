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

// MissionGetter fetches a mission by UUID.
type MissionGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*mission.Mission, error)
}

// OptionLister lists options belonging to a mission.
type OptionLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]option.Option, error)
}

// ShoppingLister lists shopping items belonging to a mission.
type ShoppingLister interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler exposes summary endpoints over HTTP.
type Handler struct {
	generator Generator
	missions  MissionGetter
	options   OptionLister
	shopping  ShoppingLister
	cache     *Cache
}

// NewHandler returns a Handler wired to the given dependencies.
func NewHandler(gen Generator, missions MissionGetter, options OptionLister, shopping ShoppingLister, cache *Cache) *Handler {
	return &Handler{
		generator: gen,
		missions:  missions,
		options:   options,
		shopping:  shopping,
		cache:     cache,
	}
}

// Generate handles POST /api/missions/{missionID}/summary.
// It loads the mission, options, and shopping items, calls the SummaryAgent,
// caches and returns the result.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission ID")
		return
	}

	m, err := h.missions.GetByID(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	opts, err := h.options.ListByMission(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	items, err := h.shopping.ListByMission(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	content, err := h.generator.Generate(r.Context(), *m, opts, items)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "summary service unavailable")
		return
	}

	s := &ShoppingSummary{
		ID:          uuid.New(),
		MissionID:   missionID,
		Content:     content,
		GeneratedAt: time.Now().UTC(),
	}

	h.cache.Set(missionID.String(), s)
	httputil.WriteJSON(w, http.StatusCreated, s)
}

// Get handles GET /api/missions/{missionID}/summary.
// Returns the cached summary if available, 404 otherwise (use POST to generate).
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission ID")
		return
	}

	cached, ok := h.cache.Get(missionID.String())
	if !ok {
		httputil.WriteError(w, http.StatusNotFound, "no summary available — POST to generate one")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, cached)
}
