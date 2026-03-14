package pricing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
)

// Handler exposes the PricingAgent as an HTTP endpoint.
type Handler struct {
	agent      *Agent
	missionSvc *mission.Service
	optionRepo option.Repository
}

// NewHandler creates a new pricing handler.
func NewHandler(agent *Agent, missionSvc *mission.Service, optionRepo option.Repository) *Handler {
	return &Handler{agent: agent, missionSvc: missionSvc, optionRepo: optionRepo}
}

// Routes mounts pricing routes under a parent router.
// Expects chi URL param "missionID" (UUID) from the parent.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.runPricing)
	return r
}

// runPricing triggers the PricingAgent for the given mission.
// POST /api/missions/{missionID}/pricing
// Fetches all existing options for the mission, runs pricing, and returns created shopping items.
func (h *Handler) runPricing(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	m, err := h.missionSvc.GetByID(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	opts, err := h.optionRepo.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if len(opts) == 0 {
		httputil.WriteError(w, http.StatusBadRequest, "no options found — run research first")
		return
	}

	items, err := h.agent.Run(r.Context(), *m, opts)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"count": len(items),
		"items": items,
	})
}
