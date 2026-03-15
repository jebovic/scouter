package budgetalert

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves budget alert endpoints using a direct pgx pool connection.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new Handler with the given database pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// GetAlerts handles GET /api/budget-alerts.
// It queries all active (non-archived) missions, computes spending percentage,
// and returns alerts only for missions that have crossed a threshold (≥60%).
func (h *Handler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	alerts, checked, err := h.computeAlerts(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, BudgetAlertList{
		Alerts:          alerts,
		MissionsChecked: checked,
	})
}

const budgetQuery = `
	SELECT
		m.id::text,
		m.name,
		m.budget::float8,
		COALESCE(SUM(si.price), 0)::float8 AS spent
	FROM missions m
	LEFT JOIN shopping_items si ON si.mission_id = m.id
	WHERE m.archived_at IS NULL
	GROUP BY m.id, m.name, m.budget
`

func (h *Handler) computeAlerts(ctx context.Context) ([]BudgetAlert, int, error) {
	rows, err := h.pool.Query(ctx, budgetQuery)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var alerts []BudgetAlert
	checked := 0

	for rows.Next() {
		var (
			id     string
			name   string
			budget float64
			spent  float64
		)
		if err := rows.Scan(&id, &name, &budget, &spent); err != nil {
			return nil, 0, err
		}
		checked++

		// Skip missions with zero budget to avoid division by zero.
		if budget <= 0 {
			continue
		}

		pct := (spent / budget) * 100

		alert, ok := buildAlert(id, name, budget, spent, pct)
		if !ok {
			continue
		}
		alerts = append(alerts, alert)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return alerts, checked, nil
}

// buildAlert returns an alert and true if the mission has crossed a threshold.
// Returns the zero value and false for missions below 60%.
func buildAlert(id, name string, budget, spent, pct float64) (BudgetAlert, bool) {
	over := spent - budget

	switch {
	case pct > 100:
		return BudgetAlert{
			MissionID:   id,
			MissionName: name,
			TotalBudget: budget,
			TotalSpent:  spent,
			Percentage:  pct,
			AlertLevel:  "critical",
			Message:     fmt.Sprintf("Dépassement de budget! +%.0f€ au-dessus du budget", over),
			Recommendations: []string{
				"Retirez les articles non essentiels de votre liste de courses.",
				"Recherchez des alternatives moins chères pour réduire les dépenses.",
				"Réallouez une partie du budget d'autres missions si possible.",
			},
		}, true

	case pct >= 80:
		return BudgetAlert{
			MissionID:   id,
			MissionName: name,
			TotalBudget: budget,
			TotalSpent:  spent,
			Percentage:  pct,
			AlertLevel:  "warning",
			Message:     fmt.Sprintf("Attention: budget presque épuisé (%.0f%%)", pct),
			Recommendations: []string{
				"Priorisez les achats les plus importants avant d'en ajouter d'autres.",
				"Comparez les prix sur plusieurs enseignes pour économiser.",
			},
		}, true

	case pct >= 60:
		return BudgetAlert{
			MissionID:   id,
			MissionName: name,
			TotalBudget: budget,
			TotalSpent:  spent,
			Percentage:  pct,
			AlertLevel:  "info",
			Message:     fmt.Sprintf("Vous avez dépensé %.0f%% de votre budget", pct),
			Recommendations: []string{
				"Votre budget avance bien. Gardez un œil sur les prochaines dépenses.",
			},
		}, true

	default:
		return BudgetAlert{}, false
	}
}
