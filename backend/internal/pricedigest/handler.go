package pricedigest

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// itemRow holds a raw row from the price query before classification.
type itemRow struct {
	itemID      string
	itemName    string
	merchant    string
	missionName string
	missionSlug string
	currentPrice float64
	status      string
	oldPrice    *float64 // nil when no price_history record exists for ≥24 h ago
}

// Handler serves the price digest endpoint.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a Handler backed by the given pgx pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

const priceDigestQuery = `
SELECT
    si.id::text,
    si.name,
    si.merchant,
    si.price::float8   AS current_price,
    si.status,
    m.name             AS mission_name,
    m.slug             AS mission_slug,
    (
        SELECT ph.price::float8
        FROM   price_history ph
        WHERE  ph.item_id = si.id
          AND  ph.recorded_at <= NOW() - INTERVAL '24 hours'
        ORDER  BY ph.recorded_at DESC
        LIMIT  1
    ) AS old_price
FROM shopping_items si
JOIN missions m ON m.id = si.mission_id
WHERE m.archived_at IS NULL
ORDER BY si.name
`

// GetDigest handles GET /api/price-digest.
// It returns a 24-hour price change digest for all items in non-archived missions.
func (h *Handler) GetDigest(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), priceDigestQuery)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var it itemRow
		if err := rows.Scan(
			&it.itemID,
			&it.itemName,
			&it.merchant,
			&it.currentPrice,
			&it.status,
			&it.missionName,
			&it.missionSlug,
			&it.oldPrice,
		); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	digest := buildDigest(items)
	httputil.WriteJSON(w, http.StatusOK, digest)
}

// buildDigest classifies rows and constructs the Digest response.
func buildDigest(items []itemRow) Digest {
	var drops, rises []PriceChange
	stable := 0
	total := len(items)

	for _, it := range items {
		if it.oldPrice == nil {
			// No historical record — treat as stable (no baseline to compare).
			stable++
			continue
		}
		old := *it.oldPrice
		if old <= 0 {
			stable++
			continue
		}
		diff := it.currentPrice - old
		pct := diff / old * 100

		if math.Abs(pct) < 0.01 {
			// Less than 0.01% change — considered stable.
			stable++
			continue
		}

		pc := PriceChange{
			ItemID:       it.itemID,
			ItemName:     it.itemName,
			Merchant:     it.merchant,
			MissionName:  it.missionName,
			MissionSlug:  it.missionSlug,
			OldPrice:     old,
			NewPrice:     it.currentPrice,
			ChangeAmount: diff,
			ChangePct:    pct,
			Status:       it.status,
			IsDropping:   diff < 0,
		}

		if diff < 0 {
			drops = append(drops, pc)
		} else {
			rises = append(rises, pc)
		}
	}

	// Sort drops by largest drop percentage (most negative first).
	sort.Slice(drops, func(i, j int) bool {
		return drops[i].ChangePct < drops[j].ChangePct
	})

	// Sort rises by largest rise percentage (most positive first).
	sort.Slice(rises, func(i, j int) bool {
		return rises[i].ChangePct > rises[j].ChangePct
	})

	var biggestDrop, biggestRise *PriceChange
	if len(drops) > 0 {
		d := drops[0]
		biggestDrop = &d
	}
	if len(rises) > 0 {
		r := rises[0]
		biggestRise = &r
	}

	summary := buildSummary(len(drops), len(rises))

	return Digest{
		GeneratedAt:       time.Now().UTC(),
		TotalItemsChecked: total,
		PriceDrops:        drops,
		PriceRises:        rises,
		StableItems:       stable,
		BiggestDrop:       biggestDrop,
		BiggestRise:       biggestRise,
		Summary:           summary,
	}
}

// buildSummary produces the French summary string.
func buildSummary(drops, rises int) string {
	switch {
	case drops == 0 && rises == 0:
		return "Aucun changement de prix détecté aujourd'hui"
	case drops == 0:
		return fmt.Sprintf("%d hausse(s) de prix détectée(s) aujourd'hui", rises)
	case rises == 0:
		return fmt.Sprintf("%d baisse(s) de prix détectée(s) aujourd'hui", drops)
	default:
		return fmt.Sprintf("%d baisse(s) et %d hausse(s) de prix détectées aujourd'hui", drops, rises)
	}
}

