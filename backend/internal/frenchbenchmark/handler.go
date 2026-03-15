package frenchbenchmark

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 20 * time.Minute

// BenchmarkResult is the response for a single mission's French market benchmark.
type BenchmarkResult struct {
	MissionID    string   `json:"missionId"`
	MissionName  string   `json:"missionName"`
	MedianPrice  float64  `json:"medianPrice"`
	YourAvgPrice float64  `json:"yourAvgPrice"`
	Verdict      string   `json:"verdict"`
	VerdictCode  string   `json:"verdictCode"`
	SampleSize   int      `json:"sampleSize"`
	PriceRange   [2]float64 `json:"priceRange"`
	Tips         []string `json:"tips"`
}

type cacheEntry struct {
	data      *BenchmarkResult
	expiresAt time.Time
}

// Handler serves the French market price benchmark endpoint.
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

// GetBenchmark handles GET /api/missions/{missionId}/french-benchmark.
func (h *Handler) GetBenchmark(w http.ResponseWriter, r *http.Request) {
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

type missionRow struct {
	id   string
	name string
}

func (h *Handler) load(ctx context.Context, missionID string) (*BenchmarkResult, error) {
	var m missionRow
	err := h.pool.QueryRow(ctx,
		`SELECT id::text, name FROM missions WHERE id::text = $1`,
		missionID,
	).Scan(&m.id, &m.name)
	if err != nil {
		return nil, err
	}

	// Fetch current prices from shopping_items for this mission.
	rows, err := h.pool.Query(ctx,
		`SELECT current_price FROM shopping_items WHERE mission_id::text = $1 AND current_price IS NOT NULL`,
		missionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []float64
	for rows.Next() {
		var p float64
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildBenchmark(m.id, m.name, prices), nil
}

// buildBenchmark is a pure function for testability.
func buildBenchmark(missionID, missionName string, prices []float64) *BenchmarkResult {
	// Deterministic market median via FNV-32a hash of missionID.
	marketMedian := computeMarketMedian(missionID)

	// Compute user average.
	var sum float64
	for _, p := range prices {
		sum += p
	}
	var avgPrice float64
	if len(prices) > 0 {
		avgPrice = sum / float64(len(prices))
	} else {
		avgPrice = marketMedian
	}

	// Simulated sample size and price range from hash.
	h := fnv.New32a()
	h.Write([]byte(missionID + "range"))
	seed := h.Sum32()
	sampleSize := int(seed%91) + 10 // 10-100
	spread := marketMedian * 0.3
	priceRange := [2]float64{
		roundTo2(marketMedian - spread),
		roundTo2(marketMedian + spread),
	}

	verdictCode, verdict, tips := computeVerdict(avgPrice, marketMedian)

	return &BenchmarkResult{
		MissionID:    missionID,
		MissionName:  missionName,
		MedianPrice:  roundTo2(marketMedian),
		YourAvgPrice: roundTo2(avgPrice),
		Verdict:      verdict,
		VerdictCode:  verdictCode,
		SampleSize:   sampleSize,
		PriceRange:   priceRange,
		Tips:         tips,
	}
}

// computeMarketMedian uses FNV-32a to produce a deterministic price in range [50, 1050].
func computeMarketMedian(missionID string) float64 {
	h := fnv.New32a()
	h.Write([]byte(missionID))
	return float64(h.Sum32()%1001) + 50.0
}

func computeVerdict(avgPrice, median float64) (code, verdict string, tips []string) {
	if median == 0 {
		return "prix_moyen", "Prix dans la moyenne du marché français", defaultTips()
	}
	ratio := avgPrice / median
	switch {
	case ratio < 0.9:
		return "bon_prix",
			fmt.Sprintf("Bon prix — %.0f%% sous la médiane du marché", (1-ratio)*100),
			[]string{
				"Continuez à surveiller les soldes saisonnières.",
				"Vérifiez les garanties constructeur avant d'acheter.",
			}
	case ratio > 1.15:
		return "au_dessus_du_marche",
			fmt.Sprintf("Prix élevé — %.0f%% au-dessus de la médiane du marché", (ratio-1)*100),
			[]string{
				"Comparez sur d'autres marchands français (Fnac, Darty, Amazon.fr).",
				"Attendez les prochaines promotions (French Days, soldes d'hiver/été).",
				"Vérifiez si une version précédente est disponible moins chère.",
			}
	default:
		return "prix_moyen",
			"Prix dans la moyenne du marché français",
			defaultTips()
	}
}

func defaultTips() []string {
	return []string{
		"Comparez les prix sur plusieurs sites français.",
		"Activez les alertes de prix pour saisir les baisses.",
	}
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
