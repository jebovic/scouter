package purchase

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves purchase record endpoints mounted under /api/missions/{missionID}/purchase.
type Handler struct {
	svc Service
}

// NewHandler creates a purchase handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Routes returns a chi.Router for the purchase sub-resource.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Get("/", h.get)
	r.Patch("/", h.update)
	return r
}

func missionIDFromURL(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "missionID"))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	missionID, err := missionIDFromURL(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.PurchasedAt == "" || req.FinalPrice <= 0 || req.Merchant == "" {
		httputil.WriteError(w, http.StatusBadRequest, "purchasedAt, finalPrice, and merchant are required")
		return
	}

	rec, err := h.svc.Record(r.Context(), missionID, req)
	if errors.Is(err, ErrAlreadyExists) {
		httputil.WriteError(w, http.StatusConflict, "purchase already recorded for this mission")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record purchase")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, rec)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	missionID, err := missionIDFromURL(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	rec, err := h.svc.Get(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get purchase record")
		return
	}
	if rec == nil {
		httputil.WriteError(w, http.StatusNotFound, "no purchase record found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, rec)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	missionID, err := missionIDFromURL(r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rec, err := h.svc.Update(r.Context(), missionID, req)
	if errors.Is(err, ErrNotFound) {
		httputil.WriteError(w, http.StatusNotFound, "no purchase record found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update purchase record")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, rec)
}
