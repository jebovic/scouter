package forecast

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes forecast endpoints.
type Handler struct {
	agent *ForecastAgent
	repo  Repository
}

// NewHandler creates a forecast Handler.
func NewHandler(agent *ForecastAgent, repo Repository) *Handler {
	return &Handler{agent: agent, repo: repo}
}

// Routes mounts the forecast routes.
// Expected to be mounted at /api/missions/{missionID}/forecast.
func (h *Handler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/", h.Run)
		r.Get("/", h.Get)
	}
}

// Run handles POST /api/missions/{missionID}/forecast.
// Runs the ForecastAgent and returns the new ForecastResult.
func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing mission id")
		return
	}

	result, err := h.agent.Run(r.Context(), missionID)
	if err != nil {
		slog.ErrorContext(r.Context(), "forecast run failed", "mission_id", missionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, result)
}

// Get handles GET /api/missions/{missionID}/forecast.
// Returns the most recent ForecastResult for the mission, or 404 if none exists.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionID")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing mission id")
		return
	}

	result, err := h.repo.GetLatest(r.Context(), missionID)
	if errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "no forecast found — run the forecast agent first")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "get latest forecast failed", "mission_id", missionID, "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}
