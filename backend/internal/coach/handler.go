package coach

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// CoachAgentInterface defines the agent contract for testing.
type CoachAgentInterface interface {
	Run(ctx context.Context, m mission.Mission, optionsCount, itemsCount int) ([]CoachTip, error)
}

// MissionGetter interface for retrieving missions by slug.
type MissionGetter interface {
	GetBySlug(ctx context.Context, slug string) (*mission.Mission, error)
}

// Handler exposes the CoachAgent as an HTTP endpoint.
type Handler struct {
	agent        CoachAgentInterface
	missionGetter MissionGetter
	optionRepo   option.Repository
	shoppingRepo shopping.Repository
	cache        *Cache
}

// NewHandler creates a new coach handler.
func NewHandler(
	agent *Agent,
	missionSvc *mission.Service,
	optionRepo option.Repository,
	shoppingRepo shopping.Repository,
) *Handler {
	return &Handler{
		agent:         agent,
		missionGetter: missionSvc,
		optionRepo:    optionRepo,
		shoppingRepo:  shoppingRepo,
		cache:         NewCache(15 * time.Minute),
	}
}

// Routes mounts coach routes under a parent router.
// Expects chi URL param "slug" (string) from the parent.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.getCoachTips)
	return r
}

// getCoachTips generates coaching tips for a mission.
// GET /api/missions/{slug}/coach
// Returns cached response if < 15min old, otherwise generates new tips.
func (h *Handler) getCoachTips(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing mission slug")
		return
	}

	m, err := h.missionGetter.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Check cache first
	if cached := h.cache.Get(m.ID); cached != nil {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Get counts for context
	options, err := h.optionRepo.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	optionsCount := len(options)

	items, err := h.shoppingRepo.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	itemsCount := len(items)

	// Generate new tips
	tips, err := h.agent.Run(r.Context(), *m, optionsCount, itemsCount)
	if err != nil {
		degraded := CoachResponse{
			MissionID:   m.ID.String(),
			Tips:        []CoachTip{{ID: "unavailable", Category: "research", Priority: "low", Title: "Analyse temporairement indisponible", Body: "Le service de coaching est temporairement indisponible. Veuillez réessayer dans quelques instants."}},
			GeneratedAt: time.Now(),
		}
		httputil.WriteJSON(w, http.StatusOK, degraded)
		return
	}

	// Cache and return
	response := CoachResponse{
		MissionID:   m.ID.String(),
		Tips:        tips,
		GeneratedAt: time.Now(),
	}
	h.cache.Set(m.ID, response)

	httputil.WriteJSON(w, http.StatusOK, response)
}
