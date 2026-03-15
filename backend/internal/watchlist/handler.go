package watchlist

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/api/watchlist", h.list)
	r.Post("/api/watchlist", h.create)
	r.Delete("/api/watchlist/{id}", h.delete)
	r.Post("/api/watchlist/check/{itemId}", h.checkPrice)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, h.repo.List(r.Context()))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.ItemName == "" || req.BaselinePrice <= 0 || req.ThresholdPct <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "itemName, baselinePrice, thresholdPct required")
		return
	}
	entry, err := h.repo.Create(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, entry)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) checkPrice(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid itemId")
		return
	}
	var body struct {
		Price float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Price <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "price required")
		return
	}
	result := h.repo.UpdatePrice(r.Context(), itemID, body.Price)
	httputil.WriteJSON(w, http.StatusOK, result)
}
