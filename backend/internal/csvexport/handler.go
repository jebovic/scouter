package csvexport

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves CSV export for mission shopping items.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new CSV export handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// row holds the denormalised data fetched from the DB for one shopping item.
type row struct {
	name             string
	price            float64
	status           string
	merchant         string
	costCategory     string
	targetPrice      *float64
	originalEstimate *float64
	pinned           bool
	createdAt        time.Time
	missionName      string
	budget           float64
}

// ExportMissionCSV handles GET /api/missions/{id}/export/csv
func (h *Handler) ExportMissionCSV(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "id")
	missionID, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	ctx := r.Context()

	// Fetch mission slug for the filename.
	var slug string
	err = h.pool.QueryRow(ctx, `SELECT slug FROM missions WHERE id = $1`, missionID).Scan(&slug)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Load all shopping items joined with the mission.
	dbRows, err := h.pool.Query(ctx, `
		SELECT si.name, si.price, si.status, si.merchant, si.cost_category,
		       si.target_price, si.original_estimate, si.pinned, si.created_at,
		       m.name AS mission_name, m.budget
		FROM shopping_items si
		JOIN missions m ON m.id = si.mission_id
		WHERE si.mission_id = $1
		ORDER BY si.cost_category, si.name
	`, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer dbRows.Close()

	var items []row
	for dbRows.Next() {
		var item row
		if scanErr := dbRows.Scan(
			&item.name,
			&item.price,
			&item.status,
			&item.merchant,
			&item.costCategory,
			&item.targetPrice,
			&item.originalEstimate,
			&item.pinned,
			&item.createdAt,
			&item.missionName,
			&item.budget,
		); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		items = append(items, item)
	}
	if err = dbRows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	filename := fmt.Sprintf("mission-%s-items.csv", slug)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)

	cw := csv.NewWriter(w)

	// Header row.
	_ = cw.Write([]string{
		"Nom", "Prix", "Statut", "Marchand", "Catégorie",
		"Prix Cible", "Estimation Initiale", "Épinglé", "Date Ajout",
	})

	for _, item := range items {
		record := []string{
			item.name,
			strconv.FormatFloat(item.price, 'f', 2, 64),
			item.status,
			item.merchant,
			item.costCategory,
			formatOptionalFloat(item.targetPrice),
			formatOptionalFloat(item.originalEstimate),
			formatBool(item.pinned),
			item.createdAt.UTC().Format(time.RFC3339),
		}
		_ = cw.Write(record)
	}

	cw.Flush()
}

func formatOptionalFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func formatBool(v bool) string {
	if v {
		return "oui"
	}
	return "non"
}
