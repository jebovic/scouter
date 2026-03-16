package scorecard

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

// Scorecard is the full response for a mission completion scorecard.
type Scorecard struct {
	MissionID    string        `json:"missionId"`
	MissionName  string        `json:"missionName"`
	Grade        string        `json:"grade"`
	Score        int           `json:"score"`
	Summary      string        `json:"summary"`
	Achievements []Achievement `json:"achievements"`
	Dimensions   Dimensions    `json:"dimensions"`
}

// Achievement is a single earned badge.
type Achievement struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Earned      bool   `json:"earned"`
}

// Dimensions breaks down the score into sub-categories.
type Dimensions struct {
	Research  int `json:"research"`
	Pricing   int `json:"pricing"`
	Budgeting int `json:"budgeting"`
	Speed     int `json:"speed"`
}

type cacheEntry struct {
	data      *Scorecard
	expiresAt time.Time
}

// Handler serves the mission completion scorecard endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	mu    sync.Mutex
	cache map[string]cacheEntry
}

// NewHandler creates a new Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:  pool,
		cache: make(map[string]cacheEntry),
	}
}

// GetScorecard handles GET /api/missions/{missionId}/scorecard.
func (h *Handler) GetScorecard(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missionId required")
		return
	}

	h.mu.Lock()
	if entry, ok := h.cache[missionID]; ok && time.Now().Before(entry.expiresAt) {
		result := entry.data
		h.mu.Unlock()
		httputil.WriteJSON(w, http.StatusOK, result)
		return
	}
	h.mu.Unlock()

	result, err := h.load(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.mu.Lock()
	h.cache[missionID] = cacheEntry{data: result, expiresAt: time.Now().Add(cacheTTL)}
	h.mu.Unlock()

	httputil.WriteJSON(w, http.StatusOK, result)
}

type missionData struct {
	id          string
	name        string
	createdAt   time.Time
	completedAt *time.Time
}

type missionStats struct {
	optionCount   int
	priceCheckCnt int
	budget        *float64
	purchasePrice *float64
}

func (h *Handler) load(ctx context.Context, missionID string) (*Scorecard, error) {
	var m missionData
	err := h.pool.QueryRow(ctx,
		`SELECT id::text, name, created_at,
		        (SELECT purchased_at FROM purchase_records WHERE mission_id = missions.id LIMIT 1)
		 FROM missions WHERE id::text = $1 OR slug = $1`,
		missionID,
	).Scan(&m.id, &m.name, &m.createdAt, &m.completedAt)
	if err != nil {
		return nil, err
	}

	var stats missionStats

	// Count options/research depth.
	_ = h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM options WHERE mission_id::text = $1`,
		m.id,
	).Scan(&stats.optionCount)

	// Count price checks (shopping items).
	_ = h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM shopping_items WHERE mission_id::text = $1`,
		m.id,
	).Scan(&stats.priceCheckCnt)

	// Budget constraint.
	_ = h.pool.QueryRow(ctx,
		`SELECT (constraints->>'budget')::float FROM missions WHERE id::text = $1 AND constraints->>'budget' IS NOT NULL`,
		m.id,
	).Scan(&stats.budget)

	// Actual purchase price.
	_ = h.pool.QueryRow(ctx,
		`SELECT price FROM purchase_records WHERE mission_id::text = $1 LIMIT 1`,
		m.id,
	).Scan(&stats.purchasePrice)

	return buildScorecard(m, stats), nil
}

// buildScorecard is a pure function for testability.
func buildScorecard(m missionData, stats missionStats) *Scorecard {
	dims := computeDimensions(m, stats)
	score := (dims.Research + dims.Pricing + dims.Budgeting + dims.Speed) / 4
	grade := gradeFor(score)
	achievements := buildAchievements(stats, dims)
	summary := buildSummary(m.name, grade, score)

	return &Scorecard{
		MissionID:    m.id,
		MissionName:  m.name,
		Grade:        grade,
		Score:        score,
		Summary:      summary,
		Achievements: achievements,
		Dimensions:   dims,
	}
}

func computeDimensions(m missionData, stats missionStats) Dimensions {
	// Research: based on number of options evaluated.
	research := 0
	switch {
	case stats.optionCount >= 5:
		research = 100
	case stats.optionCount >= 3:
		research = 75
	case stats.optionCount >= 1:
		research = 50
	}

	// Pricing: based on number of price checks.
	pricing := 0
	switch {
	case stats.priceCheckCnt >= 5:
		pricing = 100
	case stats.priceCheckCnt >= 3:
		pricing = 75
	case stats.priceCheckCnt >= 1:
		pricing = 50
	}

	// Budgeting: bought under budget?
	budgeting := 50 // neutral if no budget set
	if stats.budget != nil && stats.purchasePrice != nil && *stats.budget > 0 {
		ratio := *stats.purchasePrice / *stats.budget
		switch {
		case ratio <= 0.9:
			budgeting = 100
		case ratio <= 1.0:
			budgeting = 75
		default:
			budgeting = 25
		}
	}

	// Speed: time from creation to purchase.
	speed := 50 // neutral if not completed
	if m.completedAt != nil {
		days := m.completedAt.Sub(m.createdAt).Hours() / 24
		switch {
		case days <= 7:
			speed = 100
		case days <= 30:
			speed = 75
		case days <= 90:
			speed = 50
		default:
			speed = 25
		}
	}

	return Dimensions{
		Research:  research,
		Pricing:   pricing,
		Budgeting: budgeting,
		Speed:     speed,
	}
}

func gradeFor(score int) string {
	switch {
	case score >= 85:
		return "A"
	case score >= 70:
		return "B"
	case score >= 50:
		return "C"
	default:
		return "D"
	}
}

func buildAchievements(stats missionStats, dims Dimensions) []Achievement {
	return []Achievement{
		{
			Title:       "Chercheur Assidu",
			Description: "Évalué au moins 5 options",
			Earned:      stats.optionCount >= 5,
		},
		{
			Title:       "Pisteur de Prix",
			Description: "Suivi des prix sur au moins 5 marchands",
			Earned:      stats.priceCheckCnt >= 5,
		},
		{
			Title:       "Maître du Budget",
			Description: "Achat réalisé sous le budget prévu",
			Earned:      dims.Budgeting >= 75,
		},
		{
			Title:       "Décision Rapide",
			Description: "Mission complétée en moins d'une semaine",
			Earned:      dims.Speed == 100,
		},
	}
}

func buildSummary(name, grade string, score int) string {
	switch grade {
	case "A":
		return fmt.Sprintf("Excellent travail sur \"%s\" ! Score %d/100 — achat optimisé.", name, score)
	case "B":
		return fmt.Sprintf("Bon achat pour \"%s\". Score %d/100 — quelques axes d'amélioration.", name, score)
	case "C":
		return fmt.Sprintf("Mission \"%s\" complétée. Score %d/100 — pensez à comparer plus la prochaine fois.", name, score)
	default:
		return fmt.Sprintf("Mission \"%s\" terminée. Score %d/100 — beaucoup de marge de progression.", name, score)
	}
}

// ignorePgxNoRows swallows pgx.ErrNoRows so optional queries don't abort.
func ignorePgxNoRows(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	return err
}

var _ = ignorePgxNoRows // suppress unused warning; used conceptually via QueryRow.Scan
