package research

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

// Handler exposes the ResearchAgent as an HTTP endpoint.
type Handler struct {
	agent      *Agent
	missionSvc *mission.Service
}

// NewHandler creates a new research handler.
func NewHandler(agent *Agent, missionSvc *mission.Service) *Handler {
	return &Handler{agent: agent, missionSvc: missionSvc}
}

// Routes mounts research routes under a parent router.
// Expects chi URL param "missionID" (UUID) from the parent.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.runResearch)
	return r
}

// runResearch triggers the ResearchAgent for the given mission.
// POST /api/missions/{missionID}/research
// Returns the list of created options on success.
func (h *Handler) runResearch(w http.ResponseWriter, r *http.Request) {
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

	var fb *FeedbackInput
	var input FeedbackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err == nil && input.Feedback != "" {
		fb = &input
	}

	options, err := h.agent.Run(r.Context(), *m, fb)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]any{
		"count": len(options),
		"items": options,
	})
}
