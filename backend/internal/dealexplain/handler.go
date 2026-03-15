package dealexplain

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// Handler exposes the deal-explain endpoint over HTTP.
type Handler struct {
	getter ItemGetter
	agent  ExplainAgentInterface
	cache  *Cache
}

// NewHandler returns a Handler wired to the given item getter, agent, and cache.
func NewHandler(getter ItemGetter, agent ExplainAgentInterface, cache *Cache) *Handler {
	return &Handler{getter: getter, agent: agent, cache: cache}
}

// Explain handles GET /api/shopping-items/{id}/explain.
// Optional query params: score (0-100), trend ("rising"|"falling"|"stable").
// Returns a DealExplanation with a deterministic fallback when the LLM is unavailable.
func (h *Handler) Explain(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Validate optional score param.
	score := 50 // default mid-range
	if raw := r.URL.Query().Get("score"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 || n > 100 {
			httputil.WriteError(w, http.StatusBadRequest, "score must be an integer between 0 and 100")
			return
		}
		score = n
	}

	trend := r.URL.Query().Get("trend")
	if trend == "" {
		trend = "stable"
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

	input := ExplainInput{
		ItemID:   rawID,
		Name:     item.Name,
		Merchant: item.Merchant,
		Price:    item.Price,
		Score:    score,
		Trend:    trend,
	}

	// Try LLM; fall back to deterministic explanation on error.
	explanation, err := h.agent.Explain(r.Context(), input)
	if err != nil {
		explanation = fallbackExplanation(rawID, score)
	}

	h.cache.Set(rawID, explanation)
	httputil.WriteJSON(w, http.StatusOK, explanation)
}

// fallbackExplanation generates a deterministic French explanation from the deal score
// when the LLM is unavailable.
func fallbackExplanation(itemID string, score int) *DealExplanation {
	var text string
	var tags []string

	switch {
	case score > 80:
		text = fmt.Sprintf("Excellente affaire avec un score de %d/100. Le prix actuel est significativement en dessous de la moyenne historique. C'est le bon moment pour acheter.", score)
		tags = []string{"excellente affaire", "prix bas", "achat recommandé"}
	case score > 60:
		text = fmt.Sprintf("Bonne affaire avec un score de %d/100. Le prix est légèrement en dessous de la moyenne. Cela représente une opportunité intéressante.", score)
		tags = []string{"bonne affaire", "prix compétitif"}
	default:
		text = fmt.Sprintf("Prix dans la moyenne avec un score de %d/100. Aucune remise significative n'est détectée pour le moment. Continuez à surveiller.", score)
		tags = []string{"prix moyen", "à surveiller"}
	}

	return &DealExplanation{
		ItemID:      itemID,
		Explanation: text,
		Tags:        tags,
		GeneratedAt: time.Now().UTC(),
	}
}
