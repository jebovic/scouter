package receipt

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// analyzeRequest is the body accepted by POST /api/missions/{slug}/receipts.
type analyzeRequest struct {
	RawText  string `json:"rawText"`
	Currency string `json:"currency"`
}

// Handler exposes receipt analysis endpoints over HTTP.
type Handler struct {
	repo  Repository
	agent AgentInterface
}

// NewHandler creates a new receipt Handler.
func NewHandler(repo Repository, agent AgentInterface) *Handler {
	return &Handler{repo: repo, agent: agent}
}

// Analyze handles POST /api/missions/{slug}/receipts.
// Body: { "rawText": "...", "currency": "EUR" }
// Returns 422 if rawText is blank.
// Calls the agent, persists the result, and returns 201 with the ReceiptAnalysis.
func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.RawText) == "" {
		httputil.WriteError(w, http.StatusUnprocessableEntity, "rawText must not be blank")
		return
	}

	currency := req.Currency
	if strings.TrimSpace(currency) == "" {
		currency = "EUR"
	}

	result, err := h.agent.Analyze(r.Context(), req.RawText)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	analysis := ReceiptAnalysis{
		MissionSlug: slug,
		RawText:     req.RawText,
		Items:       result.Items,
		Total:       result.Total,
		Currency:    currency,
	}

	saved, err := h.repo.Save(r.Context(), analysis)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, saved)
}

// List handles GET /api/missions/{slug}/receipts.
// Returns the last 10 analyses for the mission, newest first.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	analyses, err := h.repo.ListByMission(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, analyses)
}
