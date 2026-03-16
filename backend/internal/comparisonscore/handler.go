package comparisonscore

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 20 * time.Minute

type cacheEntry struct {
	report    ComparisonReport
	expiresAt time.Time
}

// Handler serves smart comparison score endpoints.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewHandler creates a new Handler with the given database pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]cacheEntry),
	}
}

// GetReport handles GET /api/missions/{missionId}/comparison-score.
// It compares shopping items within a mission across price, target-price gap, and status urgency
// to produce a composite "value score" helping users decide what to buy first.
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId is required")
		return
	}

	// Check cache
	h.mu.Lock()
	if entry, ok := h.cache[missionID]; ok && time.Now().Before(entry.expiresAt) {
		resp := entry.report
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, resp)
		return
	}
	h.mu.Unlock()

	report, err := h.compute(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Store in cache
	h.mu.Lock()
	h.cache[missionID] = cacheEntry{report: report, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, report)
}

func (h *Handler) compute(ctx context.Context, missionID string) (ComparisonReport, error) {
	report := ComparisonReport{
		MissionID:  missionID,
		ItemScores: []ItemScore{},
		TopPick:    "",
		Summary:    "",
	}

	// Query budget and UUID from mission
	var missionUUID string
	var budget float64
	err := h.pool.QueryRow(ctx, `SELECT id::text, budget FROM missions WHERE id::text = $1 OR slug = $1`, missionID).Scan(&missionUUID, &budget)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			report.Summary = "Aucun article à comparer"
			return report, nil
		}
		return report, err
	}

	// Query items, excluding 'defer' status
	rows, err := h.pool.Query(ctx, `
		SELECT id, name, price, target_price, status
		FROM shopping_items
		WHERE mission_id::text = $1 AND status != 'defer'
		ORDER BY created_at
	`, missionUUID)
	if err != nil {
		return report, err
	}
	defer rows.Close()

	var itemScores []ItemScore
	for rows.Next() {
		var id, name, status string
		var price, targetPrice *float64
		if err := rows.Scan(&id, &name, &price, &targetPrice, &status); err != nil {
			return report, err
		}

		if price == nil || *price <= 0 {
			continue
		}

		score := ItemScore{
			ItemID: id,
			Name:   name,
			Price:  *price,
		}

		// priceScore: if budget > 0, score = max(0, 100 - (price/budget)*100); else 50
		if budget > 0 {
			score.PriceScore = math.Max(0, 100-((*price/budget)*100))
		} else {
			score.PriceScore = 50
		}

		// urgencyScore: based on status
		score.UrgencyScore = getUrgencyScore(status)

		// gapScore: if targetPrice != nil and price > 0: gap = (price - targetPrice) / price * 100; score = max(0, 100 - gap); else 50
		if targetPrice != nil && *targetPrice > 0 && *price > 0 {
			gap := (*price - *targetPrice) / *price * 100
			score.GapScore = math.Max(0, 100-gap)
		} else {
			score.GapScore = 50
		}

		// compositeScore = (priceScore*0.4 + urgencyScore*0.3 + gapScore*0.3)
		score.CompositeScore = (score.PriceScore * 0.4) + (score.UrgencyScore * 0.3) + (score.GapScore * 0.3)

		itemScores = append(itemScores, score)
	}
	if err = rows.Err(); err != nil {
		return report, err
	}

	// If no items, return empty report
	if len(itemScores) == 0 {
		report.Summary = "Aucun article à comparer"
		return report, nil
	}

	// Sort by compositeScore desc
	sort.Slice(itemScores, func(i, j int) bool {
		return itemScores[i].CompositeScore > itemScores[j].CompositeScore
	})

	// Assign ranks and verdicts
	for i := range itemScores {
		itemScores[i].Rank = i + 1
		itemScores[i].Verdict = getVerdict(itemScores[i].CompositeScore)
	}

	report.ItemScores = itemScores

	// Set topPick and summary
	if len(itemScores) > 0 {
		topScore := itemScores[0]
		report.TopPick = topScore.Name
		report.Summary = fmt.Sprintf("🏆 Meilleur achat: %s (score %.0f/100)", topScore.Name, topScore.CompositeScore)
	} else {
		report.Summary = "Aucun article à comparer"
	}

	return report, nil
}

func getUrgencyScore(status string) float64 {
	switch status {
	case "flash-sale":
		return 100
	case "crisis":
		return 90
	case "watch":
		return 70
	case "preorder":
		return 60
	case "buy":
		return 50
	default:
		return 30
	}
}

func getVerdict(score float64) string {
	if score >= 80 {
		return "Excellente opportunité d'achat"
	} else if score >= 60 {
		return "Bon rapport qualité-prix"
	} else if score >= 40 {
		return "Acceptable, attendre une meilleure offre"
	}
	return "Prix élevé, négocier ou différer"
}
