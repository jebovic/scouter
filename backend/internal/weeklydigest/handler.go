package weeklydigest

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = time.Hour

// Handler serves the weekly digest endpoint with a 1h in-memory cache.
type Handler struct {
	pool      *pgxpool.Pool
	mu        sync.Mutex
	cached    *WeeklyDigest
	cachedAt  time.Time
}

// NewHandler creates a Handler backed by pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

const weeklyDigestQuery = `
WITH current_prices AS (
  SELECT DISTINCT ON (item_id) item_id, price AS current_price
  FROM price_history
  ORDER BY item_id, recorded_at DESC
),
week_ago_prices AS (
  SELECT DISTINCT ON (item_id) item_id, price AS old_price
  FROM price_history
  WHERE recorded_at < NOW() - INTERVAL '7 days'
  ORDER BY item_id, recorded_at DESC
),
all_time_lows AS (
  SELECT item_id, MIN(price) AS min_price
  FROM price_history
  GROUP BY item_id
)
SELECT
  si.id::text,
  si.name,
  si.target_price,
  m.name,
  m.slug,
  cp.current_price,
  wap.old_price,
  atl.min_price
FROM shopping_items si
JOIN missions m ON m.id = si.mission_id
JOIN current_prices cp ON cp.item_id = si.id
LEFT JOIN week_ago_prices wap ON wap.item_id = si.id
LEFT JOIN all_time_lows atl ON atl.item_id = si.id
WHERE cp.current_price IS NOT NULL
  AND m.archived_at IS NULL
`

type rawRow struct {
	itemID      string
	itemName    string
	targetPrice *float64
	missionName string
	missionSlug string
	currentPrice float64
	oldPrice    *float64
	minPrice    *float64
}

// GetWeeklyDigest handles GET /api/weekly-digest.
func (h *Handler) GetWeeklyDigest(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.cached != nil && time.Since(h.cachedAt) < cacheTTL {
		cached := h.cached
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}
	h.mu.Unlock()

	rows, err := h.pool.Query(r.Context(), weeklyDigestQuery)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var rawRows []rawRow
	for rows.Next() {
		var rr rawRow
		if err := rows.Scan(
			&rr.itemID,
			&rr.itemName,
			&rr.targetPrice,
			&rr.missionName,
			&rr.missionSlug,
			&rr.currentPrice,
			&rr.oldPrice,
			&rr.minPrice,
		); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		rawRows = append(rawRows, rr)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	digest := buildWeeklyDigest(rawRows)

	h.mu.Lock()
	h.cached = digest
	h.cachedAt = time.Now()
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, digest)
}

func buildWeeklyDigest(rows []rawRow) *WeeklyDigest {
	now := time.Now().UTC()
	weekEnd := now
	weekStart := now.AddDate(0, 0, -7)

	var bigDrops, newLows, targetHits []DigestAlert

	for _, rr := range rows {
		if rr.oldPrice == nil || *rr.oldPrice <= 0 {
			// No baseline to compare — check only target_hit and new_low relative to min
			if rr.minPrice != nil && rr.currentPrice <= *rr.minPrice+0.001 {
				newLows = append(newLows, DigestAlert{
					ItemID:      rr.itemID,
					ItemName:    rr.itemName,
					MissionName: rr.missionName,
					MissionSlug: rr.missionSlug,
					OldPrice:    rr.currentPrice,
					NewPrice:    rr.currentPrice,
					TargetPrice: rr.targetPrice,
					ChangePct:   0,
					AlertType:   "new_low",
				})
			}
			if rr.targetPrice != nil && rr.currentPrice <= *rr.targetPrice {
				targetHits = append(targetHits, DigestAlert{
					ItemID:      rr.itemID,
					ItemName:    rr.itemName,
					MissionName: rr.missionName,
					MissionSlug: rr.missionSlug,
					OldPrice:    rr.currentPrice,
					NewPrice:    rr.currentPrice,
					TargetPrice: rr.targetPrice,
					ChangePct:   0,
					AlertType:   "target_hit",
				})
			}
			continue
		}

		old := *rr.oldPrice
		changePct := (rr.currentPrice - old) / old * 100

		alert := DigestAlert{
			ItemID:      rr.itemID,
			ItemName:    rr.itemName,
			MissionName: rr.missionName,
			MissionSlug: rr.missionSlug,
			OldPrice:    old,
			NewPrice:    rr.currentPrice,
			TargetPrice: rr.targetPrice,
			ChangePct:   changePct,
		}

		if changePct < -5 {
			alert.AlertType = "big_drop"
			bigDrops = append(bigDrops, alert)
		}

		if rr.minPrice != nil && rr.currentPrice <= *rr.minPrice+0.001 {
			newLowAlert := alert
			newLowAlert.AlertType = "new_low"
			newLows = append(newLows, newLowAlert)
		}

		if rr.targetPrice != nil && rr.currentPrice <= *rr.targetPrice {
			targetHitAlert := alert
			targetHitAlert.AlertType = "target_hit"
			targetHits = append(targetHits, targetHitAlert)
		}
	}

	totalAlerts := len(bigDrops) + len(newLows) + len(targetHits)

	return &WeeklyDigest{
		WeekStart:   weekStart,
		WeekEnd:     weekEnd,
		TotalAlerts: totalAlerts,
		BigDrops:    orEmpty(bigDrops),
		NewLows:     orEmpty(newLows),
		TargetHits:  orEmpty(targetHits),
		Summary:     buildFrenchSummary(len(bigDrops), len(newLows), len(targetHits)),
		GeneratedAt: now,
	}
}

func orEmpty(s []DigestAlert) []DigestAlert {
	if s == nil {
		return []DigestAlert{}
	}
	return s
}

func buildFrenchSummary(drops, lows, hits int) string {
	total := drops + lows + hits
	if total == 0 {
		return "Aucune alerte de prix cette semaine."
	}
	parts := ""
	if drops > 0 {
		parts += fmt.Sprintf("%d grande(s) baisse(s)", drops)
	}
	if lows > 0 {
		if parts != "" {
			parts += ", "
		}
		parts += fmt.Sprintf("%d nouveau(x) plancher(s)", lows)
	}
	if hits > 0 {
		if parts != "" {
			parts += ", "
		}
		parts += fmt.Sprintf("%d objectif(s) atteint(s)", hits)
	}
	return fmt.Sprintf("Cette semaine : %s.", parts)
}
