package activityfeed

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 5 * time.Minute

type cacheEntry struct {
	feed      ActivityFeed
	expiresAt time.Time
}

// Handler serves GET /api/activity-feed.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache *cacheEntry
}

// NewHandler creates a new Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetFeed handles GET /api/activity-feed.
func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	if h.cache != nil && time.Now().Before(h.cache.expiresAt) {
		cached := h.cache.feed
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}
	h.mu.Unlock()

	ctx := r.Context()

	// ── Stats query ──────────────────────────────────────────────────────────
	var totalMissions, totalItems, activeAlerts int
	var totalBudget float64

	statsRow := h.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM missions)::int,
			(SELECT COUNT(*) FROM shopping_items)::int,
			(SELECT COUNT(*) FROM shopping_items WHERE target_price IS NOT NULL AND price <= target_price)::int,
			COALESCE((SELECT SUM(budget) FROM missions WHERE budget IS NOT NULL), 0)
	`)
	if err := statsRow.Scan(&totalMissions, &totalItems, &activeAlerts, &totalBudget); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query stats")
		return
	}

	// ── Recent missions (last 30 days) ───────────────────────────────────────
	missionRows, err := h.pool.Query(ctx, `
		SELECT id, name, slug, created_at
		FROM missions
		WHERE created_at > NOW() - INTERVAL '30 days'
		ORDER BY created_at DESC
		LIMIT 5
	`)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query missions")
		return
	}
	defer missionRows.Close()

	var activities []Activity
	for missionRows.Next() {
		var id, name, slug string
		var createdAt time.Time
		if err := missionRows.Scan(&id, &name, &slug, &createdAt); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan mission row")
			return
		}
		activities = append(activities, Activity{
			Type:        "mission_created",
			MissionID:   id,
			MissionName: name,
			MissionSlug: slug,
			Description: fmt.Sprintf("Nouvelle mission : %s", name),
			OccurredAt:  createdAt,
		})
	}
	if err := missionRows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate mission rows")
		return
	}

	// ── Recent price changes (last 7 days) ───────────────────────────────────
	priceRows, err := h.pool.Query(ctx, `
		SELECT
			ph.item_id,
			si.name,
			m.id,
			m.name AS mission_name,
			m.slug,
			ph.price,
			ph.created_at
		FROM price_history ph
		JOIN shopping_items si ON si.id = ph.item_id
		JOIN missions m ON m.id = si.mission_id
		WHERE ph.created_at > NOW() - INTERVAL '7 days'
		ORDER BY ph.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query price history")
		return
	}
	defer priceRows.Close()

	// Track previous price per item to compute delta.
	// Because rows are DESC we may encounter the same item multiple times;
	// only the first occurrence (most recent) is turned into an activity.
	seen := map[string]bool{}
	for priceRows.Next() {
		var itemID, itemName, missionID, missionName, missionSlug string
		var price float64
		var createdAt time.Time
		if err := priceRows.Scan(&itemID, &itemName, &missionID, &missionName, &missionSlug, &price, &createdAt); err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan price row")
			return
		}
		if seen[itemID] {
			continue
		}
		seen[itemID] = true

		// Lookup previous price for delta calculation.
		var prevPrice float64
		prevRow := h.pool.QueryRow(ctx, `
			SELECT price FROM price_history
			WHERE item_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT 1
		`, itemID, createdAt)
		_ = prevRow.Scan(&prevPrice) // ignore error — delta stays 0 if no prior row

		delta := price - prevPrice
		var actType, description string
		if prevPrice == 0 {
			actType = "item_added"
			description = fmt.Sprintf("Article ajouté : %s (%.2f€)", itemName, price)
		} else if delta < 0 {
			actType = "price_drop"
			description = fmt.Sprintf("Prix en baisse de %.2f€ sur %s", -delta, itemName)
		} else {
			actType = "price_rise"
			description = fmt.Sprintf("Prix en hausse de %.2f€ sur %s", delta, itemName)
		}

		activities = append(activities, Activity{
			Type:        actType,
			MissionID:   missionID,
			MissionName: missionName,
			MissionSlug: missionSlug,
			ItemID:      itemID,
			ItemName:    itemName,
			Description: description,
			Delta:       delta,
			OccurredAt:  createdAt,
		})
	}
	if err := priceRows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate price rows")
		return
	}

	// Sort all activities by occurredAt DESC and cap at 15.
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].OccurredAt.After(activities[j].OccurredAt)
	})
	if len(activities) > 15 {
		activities = activities[:15]
	}

	feed := ActivityFeed{
		Activities:    activities,
		TotalMissions: totalMissions,
		TotalItems:    totalItems,
		ActiveAlerts:  activeAlerts,
		TotalBudget:   totalBudget,
		GeneratedAt:   time.Now().UTC(),
	}

	h.mu.Lock()
	h.cache = &cacheEntry{feed: feed, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, feed)
}
