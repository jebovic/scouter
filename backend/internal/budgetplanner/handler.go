package budgetplanner

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/shopping"
)

const cacheTTL = 10 * time.Minute

type cacheEntry struct {
	plan      PurchasePlan
	expiresAt time.Time
}

// Handler serves the budget-plan endpoint.
type Handler struct {
	missionRepo mission.Repository
	shoppingRepo shopping.Repository
	cache        sync.Map
}

// NewHandler wires a Handler to the pgxpool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		missionRepo:  mission.NewRepository(pool),
		shoppingRepo: shopping.NewRepository(pool),
	}
}

// GetPlan handles GET /api/missions/{id}/budget-plan.
// The mission {id} param can be a UUID (missionID) or a slug.
func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	if idParam == "" {
		// Fallback: some routes register the param as {missionID}
		idParam = chi.URLParam(r, "missionID")
	}

	ctx := r.Context()

	var m *mission.Mission
	var err error

	// Try UUID first, fall back to slug lookup.
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

	// Serve from cache if fresh.
	cacheKey := m.ID.String()
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.plan)
			return
		}
	}

	items, err := h.loadItems(ctx, m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load items")
		return
	}

	plan := Plan(m.ID.String(), m.Budget, items)

	h.cache.Store(cacheKey, cacheEntry{plan: plan, expiresAt: time.Now().Add(cacheTTL)})

	httputil.WriteJSON(w, http.StatusOK, plan)
}

func (h *Handler) loadItems(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error) {
	return h.shoppingRepo.ListByMission(ctx, missionID)
}
