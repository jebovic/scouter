package stockalert

import (
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const (
	cacheTTL  = 30 * time.Minute
	dateKey   = "2026-01-15"
)

// itemRow holds the fields we select from shopping_items.
type itemRow struct {
	name         string
	price        float64
	status       string
	costCategory string
}

// cacheEntry wraps a cached StockStatus.
type cacheEntry struct {
	data      *StockStatus
	cachedAt  time.Time
}

// Handler exposes GET /api/missions/{missionId}/items/{itemId}/stock-status.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetStockStatus handles GET /api/missions/{missionId}/items/{itemId}/stock-status.
func (h *Handler) GetStockStatus(w http.ResponseWriter, r *http.Request) {
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

	// Serve from 30-minute cache.
	cacheKey := rawItemID
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(*cacheEntry)
		if time.Since(entry.cachedAt) < cacheTTL {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		h.cache.Delete(cacheKey)
	}

	// Fetch item from DB.
	var row itemRow
	err = h.pool.QueryRow(r.Context(),
		`SELECT name, price, status, cost_category
		 FROM shopping_items
		 WHERE id = $1 AND mission_id = $2`,
		itemID, missionID,
	).Scan(&row.name, &row.price, &row.status, &row.costCategory)
	if err != nil {
		// pgx.ErrNoRows check not needed — pgxpool.QueryRow.Scan returns the error directly.
		if err.Error() == "no rows in result set" {
			httputil.WriteError(w, http.StatusNotFound, "item not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	result := simulate(rawItemID, row)
	h.cache.Store(cacheKey, &cacheEntry{data: result, cachedAt: time.Now()})
	httputil.WriteJSON(w, http.StatusOK, result)
}

// simulate builds a deterministic StockStatus using FNV-32a hash of (itemID + dateKey).
func simulate(itemID string, row itemRow) *StockStatus {
	h := fnv.New32a()
	h.Write([]byte(itemID))
	h.Write([]byte(dateKey))
	hash := int(h.Sum32())

	bucket := hash % 10

	var (
		stockStatus        string
		stockLabel         string
		alert              string
		urgency            string
		estimatedUnitsLeft *int
		restockDays        *int
	)

	switch {
	case bucket == 0:
		// rupture (out of stock)
		stockStatus = "rupture"
		stockLabel = "Rupture de stock"
		alert = "⚠️ Article en rupture de stock — commandez chez un autre marchand"
		urgency = "critical"
		days := (hash%7 + 3)
		restockDays = &days

	case bucket <= 2:
		// stock_faible (< 5 units)
		stockStatus = "stock_faible"
		stockLabel = "Stock limité"
		alert = "⚡ Stock limité — il reste moins de 5 unités"
		urgency = "high"
		units := hash%4 + 1 // 1–4
		estimatedUnitsLeft = &units

	case bucket <= 6:
		// disponible
		stockStatus = "disponible"
		stockLabel = "Disponible"
		alert = "✅ Article disponible"
		urgency = "low"

	default:
		// stock_élevé
		stockStatus = "stock_élevé"
		stockLabel = "Large disponibilité"
		alert = "🟢 Large disponibilité"
		urgency = "low"
	}

	return &StockStatus{
		ItemID:             itemID,
		Name:               row.name,
		StockStatus:        stockStatus,
		StockLabel:         stockLabel,
		EstimatedUnitsLeft: estimatedUnitsLeft,
		RestockDays:        restockDays,
		Alert:              alert,
		Urgency:            urgency,
	}
}
