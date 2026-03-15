package health

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// MissionSlugGetter fetches a mission by its URL slug.
type MissionSlugGetter interface {
	GetBySlug(ctx context.Context, slug string) (*mission.Mission, error)
}

// OptionCounter lists options for a mission (used to count them).
type OptionCounter interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]option.Option, error)
}

// ShoppingCounter lists shopping items for a mission (used to count them).
type ShoppingCounter interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error)
}

// Handler exposes the mission health score endpoint.
type Handler struct {
	missions  MissionSlugGetter
	options   OptionCounter
	shopping  ShoppingCounter
	summarizer Summarizer
	cache     *Cache
}

// NewHandler creates a Handler wired to the given dependencies.
func NewHandler(
	missions MissionSlugGetter,
	options OptionCounter,
	shopping ShoppingCounter,
	summarizer Summarizer,
	cache *Cache,
) *Handler {
	return &Handler{
		missions:  missions,
		options:   options,
		shopping:  shopping,
		summarizer: summarizer,
		cache:     cache,
	}
}

// GetHealth handles GET /api/missions/{slug}/health.
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	m, err := h.missions.GetBySlug(r.Context(), slug)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Check cache first (keyed by mission ID).
	cacheKey := m.ID.String()
	if cached, ok := h.cache.Get(cacheKey); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Fetch option and shopping counts.
	opts, err := h.options.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	items, err := h.shopping.ListByMission(r.Context(), m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Compute budget utilization percentage.
	var budgetUsedPct float64
	if m.Budget > 0 && len(items) > 0 {
		var spent float64
		for _, item := range items {
			spent += item.Price
		}
		budgetUsedPct = (spent / m.Budget) * 100
	}

	// Compute signals and overall score.
	signals := ComputeSignals(len(opts), 0, budgetUsedPct, m.Phase, len(items))
	overall := ComputeOverallScore(signals)
	grade := ComputeGrade(overall)

	report := &HealthReport{
		MissionID:    m.ID.String(),
		OverallScore: overall,
		Grade:        grade,
		Signals:      signals,
		GeneratedAt:  time.Now().UTC(),
	}

	// Get AI summary and advice (non-blocking on LLM error — fall back to empty).
	summary, advice, err := h.summarizer.BuildHealthSummary(r.Context(), m.Name, report)
	if err == nil {
		report.Summary = summary
		report.Advice = advice
	}

	h.cache.Set(cacheKey, report)
	httputil.WriteJSON(w, http.StatusOK, report)
}
