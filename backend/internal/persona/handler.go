package persona

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes persona endpoints.
type Handler struct {
	agent *PersonaAgent
	repo  Repository
}

// NewHandler creates a persona Handler.
func NewHandler(agent *PersonaAgent, repo Repository) *Handler {
	return &Handler{agent: agent, repo: repo}
}

// Routes mounts the persona routes.
// Expected to be mounted at /api/persona.
func (h *Handler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.Get)
		r.Post("/", h.Run)
	}
}

// Get handles GET /api/persona.
// Returns the most recent persona, or 404 if none has been generated yet.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.repo.GetLatest(r.Context())
	if errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "no persona found — run the persona agent first")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "get latest persona failed", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, p)
}

// Run handles POST /api/persona.
// Runs the PersonaAgent and returns the newly generated persona.
func (h *Handler) Run(w http.ResponseWriter, r *http.Request) {
	p, err := h.agent.Run(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "persona run failed", "err", err)
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, p)
}
