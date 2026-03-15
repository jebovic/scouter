package duplicatedetector

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

const cacheTTL = 30 * time.Minute

type cacheEntry struct {
	report    DuplicateReport
	expiresAt time.Time
}

// Handler exposes the duplicate-detector endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	repo  shopping.Repository
	cache sync.Map // key: missionID string → cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
		repo: shopping.NewRepository(pool),
	}
}

// GetDuplicates handles GET /api/missions/{id}/duplicates.
func (h *Handler) GetDuplicates(w http.ResponseWriter, r *http.Request) {
	missionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	key := missionID.String()

	// Check cache.
	if v, ok := h.cache.Load(key); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.report)
			return
		}
	}

	items, err := h.repo.ListByMission(context.Background(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	groups := Detect(items)

	totalDupes := 0
	for _, g := range groups {
		totalDupes += len(g.Items)
	}

	report := DuplicateReport{
		MissionID:   missionID.String(),
		Groups:      groups,
		TotalDupes:  totalDupes,
		GeneratedAt: time.Now().UTC(),
	}

	h.cache.Store(key, cacheEntry{
		report:    report,
		expiresAt: time.Now().Add(cacheTTL),
	})

	httputil.WriteJSON(w, http.StatusOK, report)
}
