package envelope

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes the envelope CRUD endpoints.
type Handler struct {
	svc *Service
}

// NewHandler constructs a Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes registers the envelope routes onto the provided chi.Router.
func Routes(h *Handler) func(chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}/summary", h.Summary)
		r.Patch("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	}
}

// List handles GET /api/envelopes.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListAll(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list envelopes")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

// Create handles POST /api/envelopes.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	env, err := h.svc.CreateEnvelope(r.Context(), req)
	if errors.Is(err, ErrNameRequired) {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, ErrInvalidAmount) {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create envelope")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, env)
}

// Summary handles GET /api/envelopes/:id/summary.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid envelope id")
		return
	}
	summary, err := h.svc.GetSummary(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get envelope summary")
		return
	}
	if summary == nil {
		httputil.WriteError(w, http.StatusNotFound, "envelope not found")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}

// Update handles PATCH /api/envelopes/:id.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid envelope id")
		return
	}
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	env, err := h.svc.UpdateEnvelope(r.Context(), id, req)
	if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrInvalidAmount) {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "envelope not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update envelope")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, env)
}

// Delete handles DELETE /api/envelopes/:id.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid envelope id")
		return
	}
	err = h.svc.DeleteEnvelope(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "envelope not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete envelope")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
