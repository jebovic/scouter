package purchaseadvisor

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/shopping"
)

// itemGetter retrieves a shopping item by UUID.
// Tests inject a fake; production wires the shopping repository.
type itemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

// Handler exposes the purchase-advice endpoint over HTTP.
type Handler struct {
	getter  itemGetter
	advisor *PurchaseAdvisor
}

// NewHandler returns a Handler wired to the given shopping repository and LLM provider.
func NewHandler(getter itemGetter, provider llm.Provider) *Handler {
	return &Handler{
		getter:  getter,
		advisor: NewPurchaseAdvisor(provider),
	}
}

// GetAdvice handles POST /api/items/{itemID}/purchase-advice.
// Reads the item from shopping_items, invokes the LLM advisor, returns advice.
// 404 if item not found; 400 if itemID is not a valid UUID.
func (h *Handler) GetAdvice(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "itemID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	item, err := h.getter.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	// Build price history from the item's current price (lightweight proxy).
	// Full price history retrieval would require an additional repo call;
	// the advisor prompt receives it when available via AdvisorRequest.
	priceHistory := []float64{item.Price}

	var targetPrice float64
	if item.TargetPrice != nil {
		targetPrice = *item.TargetPrice
	}

	req := AdvisorRequest{
		ItemID:       rawID,
		ItemName:     item.Name,
		Category:     item.CostCategory,
		CurrentPrice: item.Price,
		TargetPrice:  targetPrice,
		Status:       item.Status,
		PriceHistory: priceHistory,
	}

	advice, err := h.advisor.Advise(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, advice)
}
