package missionprogress

import (
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

const cacheTTL = 5 * time.Minute

type cacheEntry struct {
	progress  MissionProgress
	expiresAt time.Time
}

// Handler serves GET /api/missions/{id}/progress.
type Handler struct {
	pool        *pgxpool.Pool
	missionRepo mission.Repository
	cache       sync.Map
}

// NewHandler wires a Handler to the pgxpool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:        pool,
		missionRepo: mission.NewRepository(pool),
	}
}

// itemRow holds a single row from the SQL query.
type itemRow struct {
	status      string
	price       float64
	targetPrice *float64
	name        string
	budget      float64
}

// GetProgress handles GET /api/missions/{id}/progress.
// The {id} param may be a UUID (missionID) or a slug.
func (h *Handler) GetProgress(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		idParam = chi.URLParam(r, "missionID")
	}

	ctx := r.Context()

	var m *mission.Mission
	var err error

	if missionID, parseErr := uuid.Parse(idParam); parseErr == nil {
		m, err = h.missionRepo.GetByID(ctx, missionID)
	} else {
		m, err = h.missionRepo.GetBySlug(ctx, idParam)
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Serve from cache if still fresh.
	cacheKey := m.ID.String()
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.progress)
			return
		}
	}

	rows, err := h.pool.Query(ctx, `
		SELECT si.status,
		       si.price,
		       si.target_price,
		       si.name,
		       m.budget
		FROM shopping_items si
		JOIN missions m ON m.id = si.mission_id
		WHERE si.mission_id = $1`, m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query progress")
		return
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var it itemRow
		if scanErr := rows.Scan(&it.status, &it.price, &it.targetPrice, &it.name, &it.budget); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan progress row")
			return
		}
		items = append(items, it)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate progress rows")
		return
	}

	prog := buildProgress(m.ID.String(), items)

	h.cache.Store(cacheKey, cacheEntry{progress: prog, expiresAt: time.Now().Add(cacheTTL)})

	httputil.WriteJSON(w, http.StatusOK, prog)
}

// buildProgress computes MissionProgress from raw item rows. Pure function, testable without DB.
func buildProgress(missionID string, items []itemRow) MissionProgress {
	total := len(items)

	var bought, watching, deferred, flashSale int
	var budgetUsed, budgetTotal float64

	var bestSaving float64
	var topSaving *ItemSaving

	for _, it := range items {
		// Capture budget from the first row (same value for all rows of the mission).
		if budgetTotal == 0 && it.budget > 0 {
			budgetTotal = it.budget
		}

		switch it.status {
		case "buy":
			bought++
			budgetUsed += it.price
		case "watch":
			watching++
		case "defer":
			deferred++
		case "flash-sale":
			flashSale++
		}

		// Evaluate top saving: only bought items with a target_price set.
		if it.status == "buy" && it.targetPrice != nil {
			saving := *it.targetPrice - it.price
			if saving > bestSaving {
				bestSaving = saving
				topSaving = &ItemSaving{
					ItemName:     it.name,
					CurrentPrice: it.price,
					TargetPrice:  *it.targetPrice,
					Saving:       saving,
				}
			}
		}
	}

	var completionPct, budgetPct float64
	if total > 0 {
		completionPct = float64(bought) / float64(total) * 100
	}
	if budgetTotal > 0 {
		budgetPct = budgetUsed / budgetTotal * 100
	}

	return MissionProgress{
		MissionID:      missionID,
		TotalItems:     total,
		BoughtItems:    bought,
		WatchingItems:  watching,
		DeferredItems:  deferred,
		FlashSaleItems: flashSale,
		CompletionPct:  completionPct,
		BudgetUsed:     budgetUsed,
		BudgetTotal:    budgetTotal,
		BudgetPct:      budgetPct,
		TopSaving:      topSaving,
		GeneratedAt:    time.Now().UTC(),
	}
}
