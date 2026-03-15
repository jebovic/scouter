package timingrecommender

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

type cacheEntry struct {
	resp      Response
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/items/{itemId}/timing-score.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // key: itemID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetScore handles GET /api/missions/{missionId}/items/{itemId}/timing-score.
func (h *Handler) GetScore(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "missionId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	key := itemID.String()
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
		h.cache.Delete(key)
	}

	resp, found, err := h.build(r, missionID, itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !found {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	h.cache.Store(key, cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

type itemRow struct {
	id           uuid.UUID
	name         string
	price        float64
	targetPrice  *float64
	merchant     string
	costCategory string
	status       string
}

type priceStats struct {
	minPrice *float64
	avgPrice *float64
	maxPrice *float64
	count    int
}

func (h *Handler) build(r *http.Request, missionID, itemID uuid.UUID) (Response, bool, error) {
	ctx := r.Context()

	var item itemRow
	err := h.pool.QueryRow(ctx,
		`SELECT id, name, price, target_price, merchant, cost_category, status
		 FROM shopping_items
		 WHERE id = $1 AND mission_id = $2`,
		itemID, missionID,
	).Scan(&item.id, &item.name, &item.price, &item.targetPrice, &item.merchant, &item.costCategory, &item.status)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return Response{}, false, nil
		}
		return Response{}, false, err
	}

	var stats priceStats
	_ = h.pool.QueryRow(ctx,
		`SELECT MIN(price), AVG(price), MAX(price), COUNT(*) FROM price_history WHERE item_id = $1`,
		itemID,
	).Scan(&stats.minPrice, &stats.avgPrice, &stats.maxPrice, &stats.count)

	resp := buildResponse(item, stats)
	return resp, true, nil
}

func buildResponse(item itemRow, stats priceStats) Response {
	s1, window1 := signalPriceVsAverage(item.price, stats)
	s2, window2 := signalTargetProximity(item.price, item.targetPrice)
	s3, window3 := signalSalesCalendar()
	s4, window4 := signalItemStatus(item.status)
	s5, window5 := signalPriceStability(stats)

	_ = window1 // unused individual windows
	_ = window2
	_ = window4
	_ = window5

	signals := []Signal{s1, s2, s3, s4, s5}

	composite := int(
		float64(s1.Score)*s1.Weight +
			float64(s2.Score)*s2.Weight +
			float64(s3.Score)*s3.Weight +
			float64(s4.Score)*s4.Weight +
			float64(s5.Score)*s5.Weight,
	)

	recommendation, verdict, urgency := classify(composite)

	return Response{
		ItemID:         item.id.String(),
		ItemName:       item.name,
		TimingScore:    composite,
		Recommendation: recommendation,
		Verdict:        verdict,
		Signals:        signals,
		BestBuyWindow:  window3, // always from sales calendar
		UrgencyLevel:   urgency,
	}
}

// signalPriceVsAverage — weight 0.30
func signalPriceVsAverage(current float64, stats priceStats) (Signal, string) {
	w := 0.30
	if stats.avgPrice == nil || stats.count < 1 {
		return Signal{Name: "price_vs_average", Score: 50, Weight: w, Label: "Historique insuffisant"}, ""
	}
	avg := *stats.avgPrice
	if avg == 0 {
		return Signal{Name: "price_vs_average", Score: 50, Weight: w, Label: "Historique insuffisant"}, ""
	}
	pctBelowAvg := (avg - current) / avg * 100
	score := clamp(50+int(pctBelowAvg*2), 0, 100)
	var label string
	if pctBelowAvg >= 0 {
		label = fmt.Sprintf("Prix %.0f%% sous la moyenne", math.Abs(pctBelowAvg))
	} else {
		label = fmt.Sprintf("Prix %.0f%% au-dessus de la moyenne", math.Abs(pctBelowAvg))
	}
	return Signal{Name: "price_vs_average", Score: score, Weight: w, Label: label}, ""
}

// signalTargetProximity — weight 0.25
func signalTargetProximity(current float64, target *float64) (Signal, string) {
	w := 0.25
	if target == nil {
		return Signal{Name: "target_proximity", Score: 50, Weight: w, Label: "Pas d'objectif défini"}, ""
	}
	if current == 0 {
		return Signal{Name: "target_proximity", Score: 50, Weight: w, Label: "Pas d'objectif défini"}, ""
	}
	gap := (current - *target) / current * 100
	var score int
	var label string
	switch {
	case gap <= 0:
		score = 100
		label = "Objectif de prix atteint !"
	case gap < 10:
		score = 80
		label = fmt.Sprintf("À %.0f%% de l'objectif", gap)
	case gap < 25:
		score = 50
		label = fmt.Sprintf("%.0f%% de l'objectif", gap)
	default:
		score = 20
		label = fmt.Sprintf("Encore %.0f%% de l'objectif", gap)
	}
	return Signal{Name: "target_proximity", Score: score, Weight: w, Label: label}, ""
}

// signalSalesCalendar — weight 0.25 — returns (Signal, bestBuyWindow)
func signalSalesCalendar() (Signal, string) {
	w := 0.25
	month := time.Now().Month()
	type entry struct {
		score  int
		window string
	}
	months := map[time.Month]entry{
		time.January:   {30, "Soldes d'hiver (jan)"},
		time.February:  {40, "Intersaison"},
		time.March:     {45, "Printemps"},
		time.April:     {50, "Printemps"},
		time.May:       {55, "Avant soldes d'été"},
		time.June:      {85, "Soldes d'été (juin-juil)"},
		time.July:      {90, "Soldes d'été (juillet)"},
		time.August:    {40, "Fin de soldes"},
		time.September: {50, "Rentrée"},
		time.October:   {55, "Avant Black Friday"},
		time.November:  {80, "Black Friday (nov)"},
		time.December:  {75, "Fêtes de fin d'année"},
	}
	e, ok := months[month]
	if !ok {
		e = entry{50, "Période standard"}
	}
	return Signal{Name: "sales_calendar", Score: e.score, Weight: w, Label: e.window}, e.window
}

// signalItemStatus — weight 0.10
func signalItemStatus(status string) (Signal, string) {
	w := 0.10
	type entry struct {
		score int
		label string
	}
	statuses := map[string]entry{
		"flash-sale": {95, "Vente flash en cours !"},
		"crisis":     {90, "Besoin urgent"},
		"buy":        {75, "Décision d'achat prise"},
		"preorder":   {60, "En précommande"},
		"watch":      {50, "En surveillance"},
		"defer":      {20, "Achat différé"},
	}
	e, ok := statuses[status]
	if !ok {
		e = entry{50, "Statut standard"}
	}
	return Signal{Name: "item_status", Score: e.score, Weight: w, Label: e.label}, ""
}

// signalPriceStability — weight 0.10
func signalPriceStability(stats priceStats) (Signal, string) {
	w := 0.10
	if stats.count < 3 || stats.avgPrice == nil || stats.minPrice == nil || stats.maxPrice == nil {
		return Signal{Name: "price_stability", Score: 50, Weight: w, Label: "Historique insuffisant"}, ""
	}
	avg := *stats.avgPrice
	if avg == 0 {
		return Signal{Name: "price_stability", Score: 50, Weight: w, Label: "Historique insuffisant"}, ""
	}
	spread := (*stats.maxPrice - *stats.minPrice) / avg
	var score int
	var label string
	switch {
	case spread < 0.05:
		score = 70
		label = "Prix très stable"
	case spread < 0.15:
		score = 60
		label = "Prix modérément stable"
	case spread < 0.30:
		score = 50
		label = "Prix variable"
	default:
		score = 80
		label = "Prix très volatil — guetter les creux"
	}
	return Signal{Name: "price_stability", Score: score, Weight: w, Label: label}, ""
}

func classify(score int) (recommendation, verdict, urgency string) {
	switch {
	case score >= 75:
		return "buy_now", "Moment idéal pour acheter !", "high"
	case score >= 55:
		return "watch", "Conditions favorables — restez vigilant", "medium"
	default:
		return "wait", "Attendez une meilleure opportunité", "low"
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
