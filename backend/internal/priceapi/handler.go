package priceapi

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
)

// searcher is the interface the Handler uses to fetch live prices.
// The concrete *Aggregator satisfies it via aggregatorAdapter.
type searcher interface {
	Search(ctx context.Context, productName, category, locale string) (*AggregatedResult, error)
}

// aggregatorAdapter wraps *Aggregator so that it satisfies the searcher interface.
// It exists so that tests can supply a simple stub without constructing a full Aggregator.
type aggregatorAdapter struct {
	s searcher
}

func (a *aggregatorAdapter) Search(ctx context.Context, productName, category, locale string) (*AggregatedResult, error) {
	return a.s.Search(ctx, productName, category, locale)
}

// Handler exposes the live-prices endpoint.
type Handler struct {
	aggregator  searcher
	optionRepo  option.Repository
	missionRepo mission.Repository
}

// NewHandler creates a Handler wired to the given Aggregator and repositories.
func NewHandler(agg *Aggregator, optRepo option.Repository, missionRepo mission.Repository) *Handler {
	return &Handler{
		aggregator:  agg,
		optionRepo:  optRepo,
		missionRepo: missionRepo,
	}
}

// LivePrices handles GET /api/missions/{missionID}/options/{optionID}/live-prices.
// It loads the option and its parent mission, validates ownership, then fans out
// to all supporting price providers via the Aggregator.
func (h *Handler) LivePrices(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	optionID, err := uuid.Parse(chi.URLParam(r, "optionID"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option id")
		return
	}

	opt, err := h.optionRepo.GetByID(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if opt == nil {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}

	m, err := h.missionRepo.GetByID(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Validate that the option belongs to the requested mission.
	if opt.MissionID != missionID {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}

	result, err := h.aggregator.Search(r.Context(), opt.Name, m.Category, m.Locale)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}
