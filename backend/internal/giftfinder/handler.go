package giftfinder

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = 30 * time.Minute

type cacheEntry struct {
	result    GiftResult
	expiresAt time.Time
}

// Handler serves GET /api/missions/{id}/gift-suggestions.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map
}

// NewHandler returns a Handler wired to the given pgxpool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetGiftSuggestions handles GET /api/missions/{id}/gift-suggestions.
// Query parameters:
//   - occasion  — e.g. "Noël"
//   - recipient — e.g. "Lui"
//   - maxPrice  — float, defaults to mission budget
func (h *Handler) GetGiftSuggestions(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	ctx := r.Context()

	budget, missionID, err := h.loadMissionBudget(ctx, idParam)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	q := r.URL.Query()
	occasion := strings.TrimSpace(q.Get("occasion"))
	recipient := strings.TrimSpace(q.Get("recipient"))

	maxPrice := budget
	if raw := strings.TrimSpace(q.Get("maxPrice")); raw != "" {
		if parsed, parseErr := strconv.ParseFloat(raw, 64); parseErr == nil && parsed > 0 {
			maxPrice = parsed
		}
	}

	cacheKey := missionID + "|" + occasion + "|" + recipient + "|" + strconv.FormatFloat(maxPrice, 'f', 2, 64)
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.result)
			return
		}
	}

	missionName := h.loadMissionName(ctx, missionID)

	suggestions := filter(catalog, occasion, recipient, maxPrice, missionName)

	result := GiftResult{
		MissionID:   missionID,
		Budget:      budget,
		Suggestions: suggestions,
		TotalFound:  len(suggestions),
		GeneratedAt: time.Now().UTC(),
	}

	h.cache.Store(cacheKey, cacheEntry{result: result, expiresAt: time.Now().Add(cacheTTL)})

	httputil.WriteJSON(w, http.StatusOK, result)
}

// loadMissionBudget queries the mission budget and UUID by slug or UUID param.
func (h *Handler) loadMissionBudget(ctx context.Context, idParam string) (budget float64, missionID string, err error) {
	var query string
	var arg interface{}

	if _, parseErr := uuid.Parse(idParam); parseErr == nil {
		query = `SELECT id::text, budget FROM missions WHERE id = $1`
		arg = idParam
	} else {
		query = `SELECT id::text, budget FROM missions WHERE slug = $1`
		arg = idParam
	}

	row := h.pool.QueryRow(ctx, query, arg)
	err = row.Scan(&missionID, &budget)
	return budget, missionID, err
}

// loadMissionName fetches the mission name for relevance scoring; returns empty on error.
func (h *Handler) loadMissionName(ctx context.Context, missionID string) string {
	var name string
	_ = h.pool.QueryRow(ctx, `SELECT name FROM missions WHERE id = $1`, missionID).Scan(&name)
	return name
}

// filter applies occasion / recipient / maxPrice filters then scores by mission name keywords.
func filter(items []catalogItem, occasion, recipient string, maxPrice float64, missionName string) []GiftSuggestion {
	missionWords := strings.Fields(strings.ToLower(missionName))

	var matches []GiftSuggestion
	for _, item := range items {
		if maxPrice > 0 && item.estimatedPrice > maxPrice {
			continue
		}
		if occasion != "" && !containsStr(item.occasions, occasion) {
			continue
		}
		if recipient != "" && !containsStr(item.recipients, recipient) {
			continue
		}

		score := relevanceScore(item, missionWords)

		// Pick the first matching occasion for display.
		displayOccasion := ""
		if len(item.occasions) > 0 {
			displayOccasion = item.occasions[0]
		}

		// Pick the first matching recipient for display.
		displayRecipient := ""
		if len(item.recipients) > 0 {
			displayRecipient = item.recipients[0]
		}

		matches = append(matches, GiftSuggestion{
			ID:             item.id,
			Name:           item.name,
			Category:       item.category,
			PriceRange:     item.priceRange,
			EstimatedPrice: item.estimatedPrice,
			Description:    item.description,
			Occasion:       displayOccasion,
			Recipient:      displayRecipient,
			Platforms:      item.platforms,
			Score:          score,
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > 10 {
		matches = matches[:10]
	}
	return matches
}

// relevanceScore counts how many of the item's keywords appear in the mission name words.
func relevanceScore(item catalogItem, missionWords []string) int {
	if len(missionWords) == 0 {
		return 0
	}
	score := 0
	for _, kw := range item.keywords {
		for _, mw := range missionWords {
			if strings.Contains(mw, kw) || strings.Contains(kw, mw) {
				score++
				break
			}
		}
	}
	return score
}

// containsStr reports whether slice contains s (case-insensitive).
func containsStr(slice []string, s string) bool {
	lower := strings.ToLower(s)
	for _, v := range slice {
		if strings.ToLower(v) == lower {
			return true
		}
	}
	return false
}
