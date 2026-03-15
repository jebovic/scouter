package digest

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

// Repo is the interface the Handler uses to fetch price drop data.
// It is satisfied by *Repository and any test stub.
type Repo interface {
	GetPriceDrops(ctx context.Context, days int) ([]PriceDropItem, error)
}

// Handler serves the digest endpoints.
type Handler struct {
	repo Repo
}

// NewHandler creates a Handler wired to repo.
func NewHandler(repo Repo) *Handler {
	return &Handler{repo: repo}
}

// Weekly handles GET /api/digest/weekly?days=7&currency=EUR.
//
// Query params:
//   - days     int  1-30, default 7
//   - currency string, default "EUR"
func (h *Handler) Weekly(w http.ResponseWriter, r *http.Request) {
	days := 7
	if raw := r.URL.Query().Get("days"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 30 {
			httputil.WriteError(w, http.StatusBadRequest, "days must be an integer between 1 and 30")
			return
		}
		days = v
	}

	currency := strings.ToUpper(r.URL.Query().Get("currency"))
	if currency == "" {
		currency = "EUR"
	}

	rawItems, err := h.repo.GetPriceDrops(r.Context(), days)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Compute derived fields and accumulate totals.
	computed := make([]PriceDropItem, 0, len(rawItems))
	var totalSavings float64
	for _, it := range rawItems {
		drop := (it.PriceBefore - it.PriceNow) / it.PriceBefore * 100
		totalSavings += it.PriceBefore - it.PriceNow
		computed = append(computed, PriceDropItem{
			MissionID:   it.MissionID,
			MissionName: it.MissionName,
			MissionSlug: it.MissionSlug,
			ItemID:      it.ItemID,
			ItemName:    it.ItemName,
			Merchant:    it.Merchant,
			PriceBefore: it.PriceBefore,
			PriceNow:    it.PriceNow,
			DropPct:     drop,
			Currency:    it.Currency,
		})
	}

	digest := WeeklyDigest{
		GeneratedAt:  time.Now().UTC(),
		PeriodDays:   days,
		TotalDrops:   len(computed),
		TotalSavings: totalSavings,
		Currency:     currency,
		Items:        computed,
	}

	httputil.WriteJSON(w, http.StatusOK, digest)
}
