package dealaggregator

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

// sources lists the French deal aggregator sites used for mock deal generation.
var sources = []string{
	"Dealabs",
	"BonPlan",
	"CDiscount Promos",
	"Fnac Outlet",
	"Amazon Warehouse",
}

// cacheEntry holds a cached response with its creation time.
type cacheEntry struct {
	resp      *Response
	createdAt time.Time
}

// Handler exposes the deal aggregator endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetDeals handles GET /api/missions/{missionId}/items/{itemId}/deals.
func (h *Handler) GetDeals(w http.ResponseWriter, r *http.Request) {
	rawMissionID := chi.URLParam(r, "missionId")
	rawItemID := chi.URLParam(r, "itemId")

	missionID, err := uuid.Parse(rawMissionID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from 30-minute in-memory cache.
	if v, ok := h.cache.Load(rawItemID); ok {
		entry := v.(*cacheEntry)
		if time.Since(entry.createdAt) < cacheTTL {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(rawItemID)
	}

	// Fetch item from DB, verifying it belongs to the given mission.
	var name string
	var price float64
	var costCategory string

	err = h.pool.QueryRow(r.Context(),
		`SELECT name, price, cost_category FROM shopping_items WHERE id = $1 AND mission_id = $2`,
		itemID, missionID,
	).Scan(&name, &price, &costCategory)

	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	resp := generate(rawItemID, name, price)
	h.cache.Store(rawItemID, &cacheEntry{resp: resp, createdAt: time.Now()})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// generate builds a deterministic Response for the given item using FNV-32a hashing.
func generate(itemID, name string, price float64) *Response {
	// Determine how many deals to generate: 3–5 based on hash of itemID.
	baseHash := fnv32a(itemID)
	count := 3 + int(baseHash%3) // 3, 4, or 5

	if count > len(sources) {
		count = len(sources)
	}

	slug := toSlug(name)

	deals := make([]Deal, 0, count)
	var bestDeal *Deal
	bestDiscount := -1

	for i := 0; i < count; i++ {
		source := sources[i]

		// Hash itemID + source for per-source determinism.
		h := fnv.New32a()
		h.Write([]byte(itemID))
		h.Write([]byte(source))
		seed := h.Sum32()

		// Discount: 10–40% (seed maps into [0,30] + 10).
		discount := 10 + int(seed%31)

		origPrice := price
		if origPrice <= 0 {
			origPrice = 100.0
		}
		dealPrice := math.Round(origPrice*float64(100-discount)) / 100

		// Votes: 10–500.
		votes := 10 + int(seed%491)

		hot := votes > 200

		dealURL := fmt.Sprintf("https://www.dealabs.com/deals/%s-%d", slug, seed)

		title := fmt.Sprintf("%s — -%d%% sur %s", source, discount, name)

		deal := Deal{
			Source:    source,
			Title:     title,
			OrigPrice: math.Round(origPrice*100) / 100,
			DealPrice: dealPrice,
			Discount:  discount,
			URL:       dealURL,
			Hot:       hot,
			Votes:     votes,
		}
		deals = append(deals, deal)

		if discount > bestDiscount {
			bestDiscount = discount
			d := deal
			bestDeal = &d
		}
	}

	return &Response{
		ItemID:   itemID,
		Deals:    deals,
		BestDeal: bestDeal,
	}
}

// fnv32a returns the FNV-32a hash of s as a uint32.
func fnv32a(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// toSlug converts a product name to a URL-safe slug.
func toSlug(name string) string {
	lower := strings.ToLower(name)
	var b strings.Builder
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteRune('-')
		}
	}
	return b.String()
}
