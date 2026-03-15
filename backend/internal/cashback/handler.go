package cashback

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes the cashback CRUD endpoints over HTTP.
type Handler struct {
	repo *Repository
}

// NewHandler creates a Handler backed by the given repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// Routes mounts all cashback routes onto a chi sub-router.
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/summary", h.Summary)
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

// userRef resolves the collaborator ID from the X-Collaborator-ID header or
// the userRef query parameter. Returns an empty string if neither is provided.
func userRef(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Collaborator-ID")); v != "" {
		return v
	}
	return strings.TrimSpace(r.URL.Query().Get("userRef"))
}

// Create handles POST /api/cashback.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ref := userRef(r)
	if ref == "" {
		httputil.WriteError(w, http.StatusBadRequest, "userRef is required (header X-Collaborator-ID or query param)")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.ItemName) == "" {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "itemName is required")
		return
	}
	if strings.TrimSpace(req.Merchant) == "" {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "merchant is required")
		return
	}
	if req.Amount < 0 {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "amount must be >= 0")
		return
	}
	if req.PercentRate < 0 {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "percentRate must be >= 0")
		return
	}

	entry, err := h.repo.Create(ref, req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create entry")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, entry)
}

// List handles GET /api/cashback?userRef=xxx.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ref := userRef(r)
	if ref == "" {
		httputil.WriteError(w, http.StatusBadRequest, "userRef is required")
		return
	}

	entries, err := h.repo.List(ref)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list entries")
		return
	}

	if entries == nil {
		entries = []*Entry{}
	}
	httputil.WriteJSON(w, http.StatusOK, entries)
}

// Update handles PUT /api/cashback/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	entry, err := h.repo.Update(id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "entry not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update entry")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, entry)
}

// Delete handles DELETE /api/cashback/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "entry not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete entry")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Summary handles GET /api/cashback/summary?userRef=xxx.
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	ref := userRef(r)
	if ref == "" {
		httputil.WriteError(w, http.StatusBadRequest, "userRef is required")
		return
	}

	s, err := h.repo.Summary(ref)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to compute summary")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, s)
}
