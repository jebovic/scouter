package missionreport

import (
	"bytes"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler serves the mission markdown report endpoint.
type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler creates a new missionreport handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

// statusLabel maps internal status codes to French display labels and emoji.
var statusLabel = map[string]struct {
	emoji string
	label string
}{
	"flash-sale":  {"🔥", "Ventes flash"},
	"buy":         {"✅", "À acheter"},
	"watch":       {"👀", "En surveillance"},
	"preorder":    {"⏳", "Précommande"},
	"defer":       {"⏸", "Différés"},
	"crisis":      {"🚨", "Urgents"},
	"recommended": {"💡", "Recommandés"},
}

// statusOrder defines the section ordering in the report.
var statusOrder = []string{"flash-sale", "buy", "watch", "preorder", "defer", "crisis", "recommended"}

// GetReport handles GET /api/missions/{missionID}/report
func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	rawID := chi.URLParam(r, "missionID")
	missionID, err := uuid.Parse(rawID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
		return
	}

	ctx := r.Context()

	// 1. Fetch mission info.
	var m missionRow
	err = h.pool.QueryRow(ctx,
		`SELECT name, budget, slug FROM missions WHERE id = $1`,
		missionID,
	).Scan(&m.name, &m.budget, &m.slug)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// 2. Fetch shopping items.
	rows, err := h.pool.Query(ctx,
		`SELECT name, price, target_price, status, merchant, cost_category
		 FROM shopping_items
		 WHERE mission_id = $1
		 ORDER BY status, price DESC`,
		missionID,
	)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	defer rows.Close()

	var items []itemRow
	for rows.Next() {
		var item itemRow
		if scanErr := rows.Scan(
			&item.name,
			&item.price,
			&item.targetPrice,
			&item.status,
			&item.merchant,
			&item.costCategory,
		); scanErr != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// 3. Count price history checks.
	var priceCheckCount int
	err = h.pool.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM price_history ph
		 JOIN shopping_items si ON ph.item_id = si.id
		 WHERE si.mission_id = $1`,
		missionID,
	).Scan(&priceCheckCount)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// 4. Generate the report.
	report := generateReport(m, items, priceCheckCount)

	// 5. Set headers and write.
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("scouter-%s-%s.md", m.slug, today)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(report)
}

func generateReport(m missionRow, items []itemRow, priceCheckCount int) []byte {
	today := time.Now().Format("02/01/2006")

	var totalEstimated float64
	for _, item := range items {
		totalEstimated += item.price
	}

	// Group items by status.
	byStatus := make(map[string][]itemRow)
	for _, item := range items {
		byStatus[item.status] = append(byStatus[item.status], item)
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "# Rapport de mission Scouter — %s\n", m.name)
	fmt.Fprintf(&buf, "Date: %s\n\n", today)

	fmt.Fprintf(&buf, "## Résumé\n")
	fmt.Fprintf(&buf, "- Budget: %.2f€\n", m.budget)
	fmt.Fprintf(&buf, "- Nombre d'articles: %d\n", len(items))
	fmt.Fprintf(&buf, "- Total estimé: %.2f€\n", totalEstimated)
	fmt.Fprintf(&buf, "- Vérifications de prix: %d\n\n", priceCheckCount)

	fmt.Fprintf(&buf, "## Articles par statut\n\n")

	for _, status := range statusOrder {
		statusItems, ok := byStatus[status]
		if !ok || len(statusItems) == 0 {
			continue
		}
		meta, hasMeta := statusLabel[status]
		if hasMeta {
			fmt.Fprintf(&buf, "### %s %s\n\n", meta.emoji, meta.label)
		} else {
			fmt.Fprintf(&buf, "### %s\n\n", status)
		}
		for _, item := range statusItems {
			line := fmt.Sprintf("- **%s**", item.name)
			line += fmt.Sprintf(" — %.2f€", item.price)
			if item.targetPrice != nil {
				line += fmt.Sprintf(" (objectif: %.2f€)", *item.targetPrice)
			}
			if item.merchant != "" {
				line += fmt.Sprintf(" @ %s", item.merchant)
			}
			if item.costCategory != "" {
				line += fmt.Sprintf(" [%s]", item.costCategory)
			}
			fmt.Fprintln(&buf, line)
		}
		fmt.Fprintln(&buf)
	}

	// Handle any statuses not in statusOrder.
	knownStatuses := make(map[string]bool)
	for _, s := range statusOrder {
		knownStatuses[s] = true
	}
	var unknownStatuses []string
	for s := range byStatus {
		if !knownStatuses[s] {
			unknownStatuses = append(unknownStatuses, s)
		}
	}
	sort.Strings(unknownStatuses)
	for _, status := range unknownStatuses {
		statusItems := byStatus[status]
		fmt.Fprintf(&buf, "### %s\n\n", status)
		for _, item := range statusItems {
			fmt.Fprintf(&buf, "- **%s** — %.2f€\n", item.name, item.price)
		}
		fmt.Fprintln(&buf)
	}

	// Top 3 savings: items where price < targetPrice.
	fmt.Fprintf(&buf, "## Top 3 des économies réalisées\n\n")
	type saving struct {
		name   string
		amount float64
	}
	var savings []saving
	for _, item := range items {
		if item.targetPrice != nil && item.price < *item.targetPrice {
			savings = append(savings, saving{name: item.name, amount: *item.targetPrice - item.price})
		}
	}
	sort.Slice(savings, func(i, j int) bool { return savings[i].amount > savings[j].amount })
	if len(savings) == 0 {
		fmt.Fprintf(&buf, "_Aucune économie détectée pour l'instant._\n\n")
	} else {
		limit := 3
		if len(savings) < limit {
			limit = len(savings)
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(&buf, "%d. **%s** — économie de %.2f€\n", i+1, savings[i].name, savings[i].amount)
		}
		fmt.Fprintln(&buf)
	}

	// Recommended next actions.
	fmt.Fprintf(&buf, "## Prochaines actions recommandées\n\n")
	var actions []string

	if len(byStatus["flash-sale"]) > 0 {
		actions = append(actions, "Vérifier les ventes flash avant expiration")
	}
	// Check for items close to target price (within 10%).
	var nearTarget int
	for _, item := range items {
		if item.targetPrice != nil && *item.targetPrice > 0 {
			gap := (item.price - *item.targetPrice) / *item.targetPrice
			if gap > 0 && gap <= 0.1 {
				nearTarget++
			}
		}
	}
	if nearTarget > 0 {
		actions = append(actions, fmt.Sprintf("Surveiller %d article(s) proche(s) de leur objectif de prix", nearTarget))
	}
	if priceCheckCount == 0 {
		actions = append(actions, "Lancer une analyse de prix pour suivre les évolutions")
	}
	if len(byStatus["defer"]) > 0 {
		actions = append(actions, fmt.Sprintf("Réévaluer %d article(s) différé(s)", len(byStatus["defer"])))
	}
	if len(byStatus["preorder"]) > 0 {
		actions = append(actions, fmt.Sprintf("Suivre la disponibilité de %d précommande(s)", len(byStatus["preorder"])))
	}
	overshoot := totalEstimated - m.budget
	if overshoot > 0 {
		actions = append(actions, fmt.Sprintf("Réduire le budget de %.2f€ pour respecter l'enveloppe", overshoot))
	}
	if len(actions) == 0 {
		actions = append(actions, "Continuer à surveiller les prix et mettre à jour les statuts")
	}

	for _, action := range actions {
		fmt.Fprintf(&buf, "- %s\n", action)
	}
	fmt.Fprintln(&buf)

	fmt.Fprintf(&buf, "---\n")
	fmt.Fprintf(&buf, "_Rapport généré par [Scouter](https://github.com/jibei/scouter) — %s_\n",
		strings.ReplaceAll(today, "/", "-"))

	return buf.Bytes()
}
