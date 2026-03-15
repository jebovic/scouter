package loyaltypoints

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes loyalty-program endpoints over HTTP.
type Handler struct {
	svc *Service
}

// NewHandler returns a Handler backed by the given Service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListPrograms handles GET /api/loyalty/programs.
func (h *Handler) ListPrograms(w http.ResponseWriter, r *http.Request) {
	programs := h.svc.ListPrograms()
	httputil.WriteJSON(w, http.StatusOK, programs)
}

// Calculate handles POST /api/loyalty/calculate.
//
// Request body: {"retailer":"Fnac","price":299.99,"programName":"Fnac+"}
// Response:     {"pointsEarned":299.99,"estimatedCashback":3.0,"notes":"..."}
func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Price <= 0 {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "price must be > 0")
		return
	}

	if req.Retailer == "" && req.ProgramName == "" {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "retailer or programName is required")
		return
	}

	resp, err := h.svc.Calculate(req)
	if err != nil {
		if errors.Is(err, ErrProgramNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "loyalty program not found for this retailer")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "calculation failed")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}
