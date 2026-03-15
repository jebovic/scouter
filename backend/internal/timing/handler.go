package timing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes the timing advice endpoints over HTTP.
type Handler struct {
	getter  MissionInfoGetter
	cache   *Cache
	advisor Advisor
}

// NewHandler returns a Handler wired to the given mission info getter, cache,
// and advisor.
func NewHandler(getter MissionInfoGetter, cache *Cache, advisor Advisor) *Handler {
	return &Handler{getter: getter, cache: cache, advisor: advisor}
}

// GetTimingAdvice handles POST /api/missions/{id}/timing-advice.
// It serves from cache when possible; otherwise calls the TimingAgent and
// caches the result for 24 h.
func (h *Handler) GetTimingAdvice(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "id")

	if cached, ok := h.cache.Get(missionID); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	name, category, budgetEur, createdAt, err := h.getter.GetMissionInfo(r.Context(), missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	advice, err := h.advisor.Advise(r.Context(), name, category, budgetEur, createdAt)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "timing service unavailable")
		return
	}
	advice.MissionID = missionID

	h.cache.Set(missionID, advice)
	httputil.WriteJSON(w, http.StatusCreated, advice)
}
