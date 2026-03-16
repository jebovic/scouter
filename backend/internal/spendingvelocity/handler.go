package spendingvelocity

import (
	"errors"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 10 * time.Minute

// Handler serves GET /api/missions/{missionId}/spending-velocity.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // maps missionID → cacheEntry
}

type cacheEntry struct {
	report    VelocityReport
	timestamp time.Time
}

// NewHandler creates a Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetReport handles GET /api/missions/{missionId}/spending-velocity.
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	missionParam := chi.URLParam(r, "missionId")
	if missionParam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId required")
		return
	}

	// Check cache (keyed by raw param — slug or uuid)
	if cached, ok := h.cache.Load(missionParam); ok {
		entry := cached.(cacheEntry)
		if time.Since(entry.timestamp) < cacheTTL {
			httputil.WriteJSON(w, http.StatusOK, entry.report)
			return
		}
	}

	ctx := r.Context()

	// Resolve mission by UUID or slug, fetch budget and creation date
	var missionUUID string
	var budget float64
	var createdAt time.Time
	err := h.pool.QueryRow(ctx,
		`SELECT id::text, budget, created_at FROM missions WHERE id::text = $1 OR slug = $1`,
		missionParam,
	).Scan(&missionUUID, &budget, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query mission")
		return
	}

	// Fetch total spent from buy items
	var totalSpent float64
	err = h.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(price), 0) FROM shopping_items WHERE mission_id::text = $1 AND status = 'buy'`,
		missionUUID,
	).Scan(&totalSpent)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query shopping items")
		return
	}

	// Build report
	report := buildReport(missionUUID, budget, createdAt, totalSpent)

	// Cache result
	h.cache.Store(missionParam, cacheEntry{
		report:    report,
		timestamp: time.Now(),
	})

	httputil.WriteJSON(w, http.StatusOK, report)
}

// buildReport computes the VelocityReport from raw data. Pure function, testable without DB.
func buildReport(missionID string, budget float64, createdAt time.Time, totalSpent float64) VelocityReport {
	daysActive := int(math.Max(1, time.Since(createdAt).Hours()/24))
	dailyVelocity := totalSpent / float64(daysActive)
	remainingBudget := budget - totalSpent
	budgetUsedPercent := 0.0
	if budget > 0 {
		budgetUsedPercent = (totalSpent / budget) * 100
	}

	var projectedDaysLeft int
	if dailyVelocity > 0 {
		projectedDaysLeft = int(math.Ceil(remainingBudget / dailyVelocity))
	} else {
		projectedDaysLeft = -1
	}

	var paceStatus, paceLabel, advice string

	if budget <= 0 {
		paceStatus = "unconfigured"
		paceLabel = "❓ Configuration"
		advice = "Budget non configuré pour cette mission."
	} else if budgetUsedPercent >= 90 {
		paceStatus = "fast"
		paceLabel = "⚠️ Rythme élevé"
		advice = "Votre budget sera épuisé rapidement. Vérifiez vos prochains achats."
	} else if budgetUsedPercent >= 60 {
		paceStatus = "moderate"
		paceLabel = "📊 Rythme modéré"
		advice = "Vous avancez bien dans votre budget."
	} else if budgetUsedPercent >= 30 {
		paceStatus = "on_track"
		paceLabel = "✅ Rythme sain"
		advice = "Votre budget est bien géré, continuez ainsi."
	} else {
		paceStatus = "slow"
		paceLabel = "🐌 Rythme lent"
		advice = "Peu de dépenses enregistrées. Ajoutez vos achats au fil de la mission."
	}

	return VelocityReport{
		MissionID:         missionID,
		TotalBudget:       math.Round(budget*100) / 100,
		TotalSpent:        math.Round(totalSpent*100) / 100,
		RemainingBudget:   math.Round(remainingBudget*100) / 100,
		DaysActive:        daysActive,
		DailyVelocity:     math.Round(dailyVelocity*100) / 100,
		ProjectedDaysLeft: projectedDaysLeft,
		PaceStatus:        paceStatus,
		PaceLabel:         paceLabel,
		Advice:            advice,
		BudgetUsedPercent: math.Round(budgetUsedPercent*100) / 100,
	}
}
