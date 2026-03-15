package admin

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves dangerous admin endpoints for data management.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler returns a Handler with direct pool access for bulk operations.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// DeleteAllData handles DELETE /api/data.
// Requires the header X-Confirm: delete-all as an explicit safety gate.
// Table names are hardcoded constants, not user input, so string concatenation is safe.
func (h *Handler) DeleteAllData(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Confirm") != "delete-all" {
		httputil.WriteError(w, http.StatusBadRequest, "missing confirmation header X-Confirm: delete-all")
		return
	}

	// Delete in dependency order (children before parents) to avoid FK violations.
	// These table names are hardcoded constants — not user input — so concatenation is safe.
	tables := []string{
		"purchase_records",
		"notifications",
		"agent_runs",
		"price_history",
		"shopping_items",
		"options",
		"decisions",
		"llm_usage",
		"missions",
	}

	for _, table := range tables {
		if _, err := h.pool.Exec(r.Context(), "DELETE FROM "+table); err != nil { //nolint:gosec // table names are hardcoded constants
			httputil.WriteError(w, http.StatusInternalServerError, "failed to delete data from "+table)
			return
		}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "all data deleted"})
}
