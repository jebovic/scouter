package prediction

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// ShoppingRepository is the subset of the shopping repository used by this handler.
type ShoppingRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
	ListPriceHistory(ctx context.Context, itemID uuid.UUID) ([]shopping.PriceSnapshot, error)
}

// Handler serves GET /api/shopping-items/{id}/prediction.
type Handler struct {
	repo ShoppingRepository
}

// NewHandler creates a prediction Handler backed by the given shopping repository.
func NewHandler(repo ShoppingRepository) *Handler {
	return &Handler{repo: repo}
}

// GetPrediction handles GET /api/shopping-items/{id}/prediction.
func (h *Handler) GetPrediction(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	snapshots, err := h.repo.ListPriceHistory(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	points := snapshotsToPoints(snapshots)
	result := Predict(id.String(), item.Price, points)

	httputil.WriteJSON(w, http.StatusOK, result)
}

// snapshotsToPoints converts ordered price snapshots into PricePoints with
// integer day indices starting at 0. Snapshots are already returned
// ASC-ordered from ListPriceHistory.
func snapshotsToPoints(snaps []shopping.PriceSnapshot) []PricePoint {
	pts := make([]PricePoint, len(snaps))
	for i, s := range snaps {
		pts[i] = PricePoint{DayIndex: float64(i), Price: s.Price}
	}
	return pts
}
