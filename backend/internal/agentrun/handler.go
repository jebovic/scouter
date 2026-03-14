package agentrun

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes agent run history as HTTP endpoints.
type Handler struct {
	repo Repository
}

// NewHandler creates a new agent run handler.
func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// Routes mounts agent run routes under a parent router.
// Expects chi URL param "missionID" (UUID) from the parent.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.listRuns)
	return r
}

// listRuns returns agent run history for a mission.
// GET /api/missions/{missionID}/agent-runs
// Optional query param: ?type=research|pricing
func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	var agentType *AgentType
	if t := r.URL.Query().Get("type"); t != "" {
		at := AgentType(t)
		if at != AgentTypeResearch && at != AgentTypePricing {
			httputil.WriteError(w, http.StatusBadRequest, "invalid agent type: must be research or pricing")
			return
		}
		agentType = &at
	}

	runs, err := h.repo.ListByMission(r.Context(), missionID, agentType)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"runs": runs,
	})
}
