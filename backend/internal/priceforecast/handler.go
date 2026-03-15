package priceforecast

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

const (
	historyLimit  = 90  // last 90 days of price history fed to the LLM
	forecastDays  = 30  // 30-day forecast horizon
)

// ItemGetter retrieves a shopping item by UUID.
type ItemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

// PriceHistoryGetter retrieves the recent price history for a shopping item.
type PriceHistoryGetter interface {
	GetPriceHistory(ctx context.Context, itemID uuid.UUID, limit int) ([]shopping.PriceHistoryPoint, error)
}

// Handler exposes the AI price forecast endpoint over HTTP.
type Handler struct {
	agent      ForecastAgentInterface
	getter     ItemGetter
	histGetter PriceHistoryGetter
	cache      *Cache
}

// NewHandler returns a Handler wired to the given agent, item getter, history getter, and cache.
func NewHandler(agent ForecastAgentInterface, getter ItemGetter, histGetter PriceHistoryGetter, cache *Cache) *Handler {
	return &Handler{
		agent:      agent,
		getter:     getter,
		histGetter: histGetter,
		cache:      cache,
	}
}

// GetForecast handles GET /api/missions/{missionID}/shopping/{itemID}/forecast.
// Fetches item details and price history, delegates to the ForecastAgent, and caches the result.
func (h *Handler) GetForecast(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from cache.
	if cached, ok := h.cache.Get(rawItemID); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Resolve item details.
	item, err := h.getter.GetByID(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	// Fetch price history.
	histPts, err := h.histGetter.GetPriceHistory(r.Context(), itemID, historyLimit)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Convert to DataPoint slice.
	historical := make([]DataPoint, len(histPts))
	for i, pt := range histPts {
		historical[i] = DataPoint{Date: pt.RecordedAt, Price: pt.Price}
	}

	req := ForecastRequest{
		ItemID:      rawItemID,
		ItemName:    item.Name,
		Currency:    "EUR",
		Historical:  historical,
		HorizonDays: forecastDays,
	}

	result, err := h.agent.Forecast(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "forecast unavailable")
		return
	}

	h.cache.Set(rawItemID, result)
	httputil.WriteJSON(w, http.StatusOK, result)
}
