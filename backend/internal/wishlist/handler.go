package wishlist

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jibei/scouter/internal/httputil"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func Routes(repo *Repository) func(chi.Router) {
	h := NewHandler(repo)
	return func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Delete("/{id}", h.Delete)
		r.Patch("/{id}/alert", h.UpdateAlert)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.repo.List(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list wish list items")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateWishListItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Currency == "" {
		req.Currency = "EUR"
	}
	item, err := h.repo.Create(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create wish list item")
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateWishListItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AlertEnabled == nil {
		httputil.WriteError(w, http.StatusBadRequest, "alertEnabled is required")
		return
	}
	err := h.repo.UpdateAlertEnabled(r.Context(), id, *req.AlertEnabled)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "wish list item not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update alert")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := h.repo.Delete(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "wish list item not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete wish list item")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
