package pricecomparison

import (
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = time.Hour

var retailers = []string{"Fnac", "Darty", "Amazon", "Cdiscount", "Boulanger", "Ldlc", "RueDuCommerce"}

type cacheEntry struct {
	resp      Response
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/items/{itemId}/price-comparison.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetComparison handles GET /api/missions/{missionId}/items/{itemId}/price-comparison.
func (h *Handler) GetComparison(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(key)
	}

	resp, found, err := h.build(r, missionID, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	h.cache.Store(key, cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) build(r *http.Request, missionID, itemID uuid.UUID) (Response, bool, error) {
	ctx := r.Context()

	var name, merchant, costCategory string
	var basePrice float64

	err := h.pool.QueryRow(ctx,
		`SELECT name, price, merchant, cost_category
		 FROM shopping_items
		 WHERE id = $1 AND mission_id = $2`,
		itemID, missionID,
	).Scan(&name, &basePrice, &merchant, &costCategory)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Response{}, false, nil
		}
		return Response{}, false, err
	}

	retailerPrices := make([]RetailerPrice, 0, len(retailers))
	for _, retailer := range retailers {
		h32 := fnv.New32a()
		_, _ = h32.Write([]byte(itemID.String() + retailer))
		hash := h32.Sum32()

		ratio := 0.85 + (float64(hash%100)/100.0)*0.30
		simulatedPrice := math.Round(basePrice*ratio*100) / 100

		diffPercent := (simulatedPrice - basePrice) / basePrice * 100

		var retailerURL string
		if strings.EqualFold(retailer, "Amazon") {
			retailerURL = "https://www.amazon.fr/s?k=" + url.QueryEscape(name)
		} else {
			retailerURL = fmt.Sprintf("https://www.%s.fr/search?q=%s",
				strings.ToLower(retailer), url.QueryEscape(name))
		}

		retailerPrices = append(retailerPrices, RetailerPrice{
			Retailer:    retailer,
			Price:       simulatedPrice,
			URL:         retailerURL,
			IsBestPrice: false,
			DiffPercent: math.Round(diffPercent*10) / 10,
		})
	}

	// Sort by price ascending.
	sort.Slice(retailerPrices, func(i, j int) bool {
		return retailerPrices[i].Price < retailerPrices[j].Price
	})

	// Mark best price.
	if len(retailerPrices) > 0 {
		retailerPrices[0].IsBestPrice = true
	}

	var bestDeal *RetailerPrice
	var maxSaving float64
	if len(retailerPrices) > 0 {
		best := retailerPrices[0]
		bestDeal = &best
		saving := basePrice - best.Price
		if saving > 0 {
			maxSaving = math.Round(saving*100) / 100
		}
	}

	return Response{
		ItemID:    itemID.String(),
		ItemName:  name,
		BasePrice: basePrice,
		Retailers: retailerPrices,
		BestDeal:  bestDeal,
		MaxSaving: maxSaving,
	}, true, nil
}
