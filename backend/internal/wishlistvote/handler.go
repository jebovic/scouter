package wishlistvote

import (
	"math"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
)

const cacheTTL = 10 * time.Minute

type cacheEntry struct {
	resp      VoteSummaryResponse
	expiresAt time.Time
}

// Handler serves GET /api/missions/{missionId}/vote-summary.
type Handler struct {
	pool        *pgxpool.Pool
	missionRepo mission.Repository
	cache       sync.Map
}

// NewHandler creates a Handler wired to the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool:        pool,
		missionRepo: mission.NewRepository(pool),
	}
}

// optionRow holds raw data from the options query.
type optionRow struct {
	id    string
	name  string
	price float64
}

// voteRow holds raw aggregated vote counts per option.
type voteRow struct {
	optionID string
	ups      int
	downs    int
}

// GetVoteSummary handles GET /api/missions/{missionId}/vote-summary.
// The {missionId} param may be a UUID or a slug.
func (h *Handler) GetVoteSummary(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "missionId")
	if idParam == "" {
		idParam = chi.URLParam(r, "id")
	}
	if idParam == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing missionId")
		return
	}

	ctx := r.Context()

	var m *mission.Mission
	var err error

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

	cacheKey := m.ID.String()
	if v, ok := h.cache.Load(cacheKey); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
	}

	// Fetch options for this mission (name + price from options table).
	// price_range is JSONB {min,max,best}; fall back to 0 if not set.
	optRows, err := h.pool.Query(ctx, `
		SELECT id, name, COALESCE((price_range->>'best')::numeric, 0) AS price
		FROM options
		WHERE mission_id = $1`, m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query options")
		return
	}
	defer optRows.Close()

	optionMap := make(map[string]optionRow)
	for optRows.Next() {
		var row optionRow
		if scanErr := optRows.Scan(&row.id, &row.name, &row.price); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan option row")
			return
		}
		optionMap[row.id] = row
	}
	if rowsErr := optRows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate option rows")
		return
	}
	optRows.Close()

	// Aggregate votes for all options in this mission.
	// vote = 1 → upvote, vote = -1 → downvote, vote = 0 → neutral (excluded from counts).
	voteRows, err := h.pool.Query(ctx, `
		SELECT option_id,
		       SUM(CASE WHEN vote = 1  THEN 1 ELSE 0 END)::int AS ups,
		       SUM(CASE WHEN vote = -1 THEN 1 ELSE 0 END)::int AS downs
		FROM option_votes
		WHERE option_id = ANY(
			SELECT id FROM options WHERE mission_id = $1
		)
		GROUP BY option_id`, m.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to query votes")
		return
	}
	defer voteRows.Close()

	voteMap := make(map[string]voteRow)
	for voteRows.Next() {
		var row voteRow
		if scanErr := voteRows.Scan(&row.optionID, &row.ups, &row.downs); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to scan vote row")
			return
		}
		voteMap[row.optionID] = row
	}
	if rowsErr := voteRows.Err(); rowsErr != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to iterate vote rows")
		return
	}
	voteRows.Close()

	resp := buildSummary(m.ID.String(), optionMap, voteMap)
	h.cache.Store(cacheKey, cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// buildSummary assembles VoteSummaryResponse from raw query data.
// Pure function — testable without DB.
func buildSummary(missionID string, optionMap map[string]optionRow, voteMap map[string]voteRow) VoteSummaryResponse {
	items := make([]VotedItem, 0, len(optionMap))

	for id, opt := range optionMap {
		vr := voteMap[id] // zero-value if no votes
		totalVotes := vr.ups + vr.downs
		netScore := vr.ups - vr.downs

		var approvalRate float64
		if totalVotes > 0 {
			approvalRate = float64(vr.ups) / float64(totalVotes)
		}

		items = append(items, VotedItem{
			ItemID:       id,
			Name:         opt.name,
			Price:        opt.price,
			Ups:          vr.ups,
			Downs:        vr.downs,
			NetScore:     netScore,
			TotalVotes:   totalVotes,
			ApprovalRate: approvalRate,
			Sentiment:    sentiment(approvalRate, totalVotes),
		})
	}

	// Sort descending by netScore, then by name for determinism.
	sort.Slice(items, func(i, j int) bool {
		if items[i].NetScore != items[j].NetScore {
			return items[i].NetScore > items[j].NetScore
		}
		return items[i].Name < items[j].Name
	})

	var mostApproved *VotedItem
	var mostControversial *VotedItem

	for idx := range items {
		it := &items[idx]
		if it.TotalVotes == 0 {
			continue
		}

		// Most approved: highest approvalRate (ties broken by totalVotes).
		if mostApproved == nil ||
			it.ApprovalRate > mostApproved.ApprovalRate ||
			(it.ApprovalRate == mostApproved.ApprovalRate && it.TotalVotes > mostApproved.TotalVotes) {
			cp := *it
			mostApproved = &cp
		}

		// Most controversial: highest totalVotes with approvalRate closest to 0.5.
		if mostControversial == nil {
			cp := *it
			mostControversial = &cp
		} else {
			itDist := math.Abs(it.ApprovalRate - 0.5)
			curDist := math.Abs(mostControversial.ApprovalRate - 0.5)
			if itDist < curDist || (itDist == curDist && it.TotalVotes > mostControversial.TotalVotes) {
				cp := *it
				mostControversial = &cp
			}
		}
	}

	return VoteSummaryResponse{
		MissionID:         missionID,
		Items:             items,
		MostApproved:      mostApproved,
		MostControversial: mostControversial,
	}
}
