package decision

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes decision endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a decision Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes mounts the decision routes under a parent router.
// Expected to be mounted at /api/missions/{missionID}/decision.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Post("/", h.Run)
	r.Get("/", h.Get)
	return r
}

// Run handles POST /api/missions/{missionID}/decision
// Triggers a full score + LLM run and returns the new Decision.
func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	d, err := h.svc.Run(r.Context(), missionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrMissionNotFound):
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
		case errors.Is(err, ErrNoOptions):
			httputil.WriteError(w, http.StatusUnprocessableEntity, "no options to score — run Research Agent first")
		default:
			slog.ErrorContext(r.Context(), "decision run failed", "mission_id", missionID, "err", err)
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, d)
}

// Get handles GET /api/missions/{missionID}/decision
// Returns the most recent Decision for the mission, or 404 if none exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	d, err := h.svc.GetLatest(r.Context(), missionID)
	if err != nil {
		slog.ErrorContext(r.Context(), "get latest decision failed", "mission_id", missionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if d == nil {
		httputil.WriteError(w, http.StatusNotFound, "no decision found — run the decision engine first")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, d)
}
