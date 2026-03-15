package vote

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves vote endpoints mounted under /api/options/{optionID}/votes.
type Handler struct {
	svc *Service
}

// NewHandler creates a new vote handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Routes returns a chi.Router for vote sub-resources.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.recordVote)
	r.Get("/", h.getAggregated)
	return r
}

func (h *Handler) recordVote(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "optionID")

	var req RecordVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CollaboratorID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "collaboratorId is required")
		return
	}
	if req.Vote < -1 || req.Vote > 1 {
		httputil.WriteError(w, http.StatusBadRequest, "vote must be -1, 0, or 1")
		return
	}

	if err := h.svc.RecordVote(r.Context(), optionID, req.CollaboratorID, req.Vote); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to record vote")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getAggregated(w http.ResponseWriter, r *http.Request) {
	optionID := chi.URLParam(r, "optionID")

	agg, err := h.svc.GetAggregated(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, agg)
}
