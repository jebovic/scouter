package competitorprice

import (
	"context"
	"hash/fnv"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// platforms lists the French e-commerce platforms with their search URL templates.
var platforms = []struct {
	name      string
	searchFmt string
}{
	{"Fnac", "https://www.fnac.com/SearchResult/ResultList.aspx?Search=%s"},
	{"Darty", "https://www.darty.com/nav/recherche?text=%s"},
	{"Cdiscount", "https://www.cdiscount.com/search/10/%s.html"},
	{"Amazon.fr", "https://www.amazon.fr/s?k=%s"},
	{"Boulanger", "https://www.boulanger.com/result?tr=%s"},
	{"Leclerc", "https://www.e.leclerc/recherche?q=%s"},
	{"Carrefour", "https://www.carrefour.fr/s?q=%s"},
	{"La Redoute", "https://www.laredoute.fr/pplp/cat=rub/search.aspx?searchtext=%s"},
}

// itemGetter retrieves a shopping item by ID.
type itemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

type pgItemGetter struct {
	pool *pgxpool.Pool
}

func newPGItemGetter(pool *pgxpool.Pool) itemGetter {
	return &pgItemGetter{pool: pool}
}

func (g *pgItemGetter) GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error) {
	return shopping.NewRepository(g.pool).GetByID(ctx, id)
}

// cacheEntry wraps a CompetitorPrices result with its generation time.
type cacheEntry struct {
	data *CompetitorPrices
}

// Handler exposes the competitor-prices endpoint.
type Handler struct {
	getter itemGetter
	cache  sync.Map // key: itemID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{getter: newPGItemGetter(pool)}
}

const cacheTTL = 30 * time.Minute

// GetCompetitorPrices handles GET /api/missions/{missionId}/items/{itemId}/competitor-prices.
func (h *Handler) GetCompetitorPrices(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from 30-minute cache.
	if v, ok := h.cache.Load(rawItemID); ok {
		entry := v.(*cacheEntry)
		if time.Since(entry.data.GeneratedAt) < cacheTTL {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		h.cache.Delete(rawItemID)
	}

	item, err := h.getter.GetByID(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	result := generate(item)
	h.cache.Store(rawItemID, &cacheEntry{data: result})
	httputil.WriteJSON(w, http.StatusOK, result)
}

// generate builds a deterministic CompetitorPrices from an item using FNV-32a hashing.
func generate(item *shopping.Item) *CompetitorPrices {
	encodedName := url.QueryEscape(item.Name)
	competitors := make([]CompetitorOffer, 0, len(platforms))

	var bestIdx int
	bestPrice := math.MaxFloat64

	for i, p := range platforms {
		// Hash itemID + platform name for determinism.
		h := fnv.New32a()
		h.Write([]byte(item.ID.String()))
		h.Write([]byte(p.name))
		seed := int(h.Sum32())

		// hash % 20 gives 0–19; map to ±10% variation: offset = (seed%20 - 10) / 100.
		variation := float64(seed%20-10) / 100.0
		price := math.Round(item.Price*(1+variation)*100) / 100

		diff := price - item.Price
		diffPct := 0.0
		if item.Price > 0 {
			diffPct = math.Round(diff/item.Price*10000) / 100
		}

		// Availability: hash % 3 → "En stock" (0,1) or "Sur commande" (2).
		availability := "En stock"
		if seed%3 == 2 {
			availability = "Sur commande"
		}

		searchURL := formatURL(p.searchFmt, encodedName)

		competitors = append(competitors, CompetitorOffer{
			Platform:     p.name,
			Price:        price,
			PriceDiff:    math.Round(diff*100) / 100,
			PriceDiffPct: diffPct,
			URL:          searchURL,
			Availability: availability,
		})

		if price < bestPrice {
			bestPrice = price
			bestIdx = i
		}
	}

	// Mark best offer.
	competitors[bestIdx].IsBest = true

	savings := math.Round((item.Price-bestPrice)*100) / 100
	if savings < 0 {
		savings = 0
	}

	best := competitors[bestIdx]
	return &CompetitorPrices{
		ItemID:      item.ID.String(),
		ItemName:    item.Name,
		BasePrice:   item.Price,
		Competitors: competitors,
		BestOffer:   &best,
		Savings:     savings,
		GeneratedAt: time.Now().UTC(),
	}
}

// formatURL inserts the encoded search term into a printf-style URL template.
func formatURL(template, encoded string) string {
	// Use simple string replacement for the single %s placeholder.
	result := make([]byte, 0, len(template)+len(encoded))
	for i := 0; i < len(template); i++ {
		if i < len(template)-1 && template[i] == '%' && template[i+1] == 's' {
			result = append(result, []byte(encoded)...)
			i++ // skip 's'
		} else {
			result = append(result, template[i])
		}
	}
	return string(result)
}
