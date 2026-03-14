package usage

import (
	"net/http"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes the usage summary API.
type Handler struct {
	svc *Service
}

// NewHandler creates a new usage Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetSummary handles GET /api/usage?period=7d
func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "7d"
	}
	summary, err := h.svc.GetSummary(r.Context(), period)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load usage summary")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}
