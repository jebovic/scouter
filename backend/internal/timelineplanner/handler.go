package timelineplanner

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

// WeekBucket represents one week in the timeline.
type WeekBucket struct {
	Week        int      `json:"week"`
	Label       string   `json:"label"`
	Items       []string `json:"items"`
	Action      string   `json:"action"`
	PromoHint   string   `json:"promoHint"`
	ItemCount   int      `json:"itemCount"`
}

// Timeline is the full response.
type Timeline struct {
	MissionID   string       `json:"missionId"`
	MissionName string       `json:"missionName"`
	Buckets     []WeekBucket `json:"buckets"`
	Summary     string       `json:"summary"`
	TotalItems  int          `json:"totalItems"`
}

type cacheEntry struct {
	data      *Timeline
	expiresAt time.Time
}

// Handler serves the purchase timeline planner endpoint.
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

// GetTimeline handles GET /api/missions/{missionId}/purchase-timeline.
func (h *Handler) GetTimeline(w http.ResponseWriter, r *http.Request) {
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

type shoppingItem struct {
	id     string
	name   string
	status string
}

func (h *Handler) load(ctx context.Context, missionID string) (*Timeline, error) {
	var missionName string
	err := h.pool.QueryRow(ctx,
		`SELECT name FROM missions WHERE id::text = $1`,
		missionID,
	).Scan(&missionName)
	if err != nil {
		return nil, err
	}

	rows, err := h.pool.Query(ctx,
		`SELECT id::text, name, COALESCE(status, 'pending') FROM shopping_items WHERE mission_id::text = $1 ORDER BY created_at`,
		missionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []shoppingItem
	for rows.Next() {
		var it shoppingItem
		if err := rows.Scan(&it.id, &it.name, &it.status); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return buildTimeline(missionID, missionName, items), nil
}

// buildTimeline is a pure function for testability.
func buildTimeline(missionID, missionName string, items []shoppingItem) *Timeline {
	// Distribute items into 4 weekly buckets by status.
	buckets := initBuckets(missionID)

	for _, item := range items {
		weekIdx := statusToWeek(item.status)
		buckets[weekIdx].Items = append(buckets[weekIdx].Items, item.name)
		buckets[weekIdx].ItemCount++
	}

	summary := buildSummary(missionName, items)

	return &Timeline{
		MissionID:   missionID,
		MissionName: missionName,
		Buckets:     buckets,
		Summary:     summary,
		TotalItems:  len(items),
	}
}

// initBuckets creates 4 weekly buckets with deterministic promo hints.
func initBuckets(missionID string) []WeekBucket {
	promos := promoHints(missionID)
	now := time.Now()
	buckets := make([]WeekBucket, 4)
	for i := 0; i < 4; i++ {
		weekStart := now.AddDate(0, 0, i*7)
		buckets[i] = WeekBucket{
			Week:      i + 1,
			Label:     fmt.Sprintf("Semaine %d — %s", i+1, weekStart.Format("02 Jan")),
			Items:     []string{},
			Action:    weekAction(i),
			PromoHint: promos[i],
			ItemCount: 0,
		}
	}
	return buckets
}

func weekAction(weekIdx int) string {
	switch weekIdx {
	case 0:
		return "Recherche & comparaison"
	case 1:
		return "Surveillance des prix"
	case 2:
		return "Décision d'achat"
	default:
		return "Achat final"
	}
}

// promoHints generates 4 deterministic promo hint strings via FNV hash.
func promoHints(missionID string) [4]string {
	hints := [...]string{
		"Vérifiez les codes promo actifs",
		"French Days possibles — restez alerté",
		"Soldes saisonnières approchent",
		"Cashback disponible sur certains sites",
		"Vente flash Fnac/Darty à surveiller",
		"Comparez avec Amazon.fr et Cdiscount",
	}

	var result [4]string
	for i := 0; i < 4; i++ {
		h := fnv.New32a()
		h.Write([]byte(fmt.Sprintf("%s:week%d", missionID, i)))
		result[i] = hints[h.Sum32()%uint32(len(hints))]
	}
	return result
}

// statusToWeek maps shopping item status to a week bucket index (0-3).
func statusToWeek(status string) int {
	switch status {
	case "purchased", "done":
		return 3
	case "watching", "shortlisted":
		return 2
	case "researching":
		return 1
	default:
		return 0
	}
}

func buildSummary(missionName string, items []shoppingItem) string {
	if len(items) == 0 {
		return fmt.Sprintf("Aucun article dans la mission \"%s\". Ajoutez des articles à suivre.", missionName)
	}
	return fmt.Sprintf("%d article(s) planifié(s) sur 4 semaines pour \"%s\".", len(items), missionName)
}
