package currency

import (
	"math"
	"net/http"
	"strconv"

	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves currency rate and conversion endpoints.
type Handler struct {
	svc *Service
}

// NewHandler returns a new Handler wrapping the given Service.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetRates handles GET /api/currency/rates.
func (h *Handler) GetRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.svc.GetRates()
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch exchange rates")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rates)
}

// Convert handles GET /api/currency/convert?from=USD&to=EUR&amount=100.
func (h *Handler) Convert(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	amountStr := q.Get("amount")

	if from == "" || to == "" || amountStr == "" {
		httputil.WriteError(w, http.StatusBadRequest, "from, to, and amount are required")
		return
	}

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "amount must be a non-negative number")
		return
	}

	result, err := h.svc.Convert(amount, from, to)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Round to 2 decimal places for display.
	result = math.Round(result*100) / 100

	httputil.WriteJSON(w, http.StatusOK, ConvertResponse{
		Result: result,
		From:   from,
		To:     to,
		Amount: amount,
	})
}
