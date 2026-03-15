package negotiate

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// ItemGetter retrieves a shopping item by UUID.
// Tests inject a fake; production wires the shopping repository.
type ItemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

// Handler exposes the negotiate endpoint over HTTP.
type Handler struct {
	getter ItemGetter
	agent  AgentInterface
	cache  *Cache
}

// NewHandler returns a Handler wired to the given item getter, agent, and cache.
func NewHandler(getter ItemGetter, agent AgentInterface, cache *Cache) *Handler {
	return &Handler{getter: getter, agent: agent, cache: cache}
}

// Get handles GET /api/shopping-items/{id}/negotiate.
// Returns a NegotiationAdvice with tips and a ready-to-use French script.
// 404 if item not found.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from cache.
	if cached, ok := h.cache.Get(rawID); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	// Resolve item details.
	item, err := h.getter.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	advice, err := h.agent.Negotiate(r.Context(), rawID, item.Name, item.Merchant, item.Price)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("negotiate agent error: %v", err))
		return
	}

	h.cache.Set(rawID, advice)
	httputil.WriteJSON(w, http.StatusOK, advice)
}
