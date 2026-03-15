package salecalendar

import (
	"net/http"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

// staticEvents holds the hard-coded 2026 French sale calendar.
// Dates are in Paris local time (CET/CEST); we use time.UTC for simplicity
// as only the calendar date matters for daysUntil / isActive.
var staticEvents = []struct {
	id            string
	name          string
	eventType     string
	start         string
	end           string
	discountRange string
	categories    []string
	tip           string
}{
	{
		id: "soldes-hiver-2026", name: "Soldes d'hiver", eventType: "soldes",
		start: "2026-01-08", end: "2026-02-18",
		discountRange: "20-70%",
		categories:    []string{"mode", "sport", "maison"},
		tip:           "Profitez des dernières semaines pour les meilleures remises — les stocks fondent !",
	},
	{
		id: "saint-valentin-2026", name: "Saint-Valentin", eventType: "fete",
		start: "2026-02-07", end: "2026-02-14",
		discountRange: "10-30%",
		categories:    []string{"bijoux", "mode", "parfum"},
		tip:           "Les cadeaux personnalisés ont souvent les meilleures promos en début de période.",
	},
	{
		id: "french-days-2026", name: "French Days", eventType: "flash",
		start: "2026-04-29", end: "2026-05-04",
		discountRange: "30-60%",
		categories:    []string{"électronique", "mode", "maison"},
		tip:           "Comparable au Black Friday : soyez prêt avec votre liste de souhaits dès le premier jour.",
	},
	{
		id: "fete-des-meres-2026", name: "Fête des Mères", eventType: "fete",
		start: "2026-05-24", end: "2026-05-31",
		discountRange: "10-25%",
		categories:    []string{"beauté", "bijoux", "fleurs"},
		tip:           "Commandez tôt pour éviter les ruptures de stock sur les coffrets beauté.",
	},
	{
		id: "soldes-ete-2026", name: "Soldes d'été", eventType: "soldes",
		start: "2026-06-25", end: "2026-07-22",
		discountRange: "20-70%",
		categories:    []string{"mode", "sport", "jardin"},
		tip:           "Les premières 48h offrent les plus grandes remises et les meilleures tailles.",
	},
	{
		id: "amazon-prime-day-2026", name: "Amazon Prime Day", eventType: "flash",
		start: "2026-07-12", end: "2026-07-13",
		discountRange: "20-50%",
		categories:    []string{"électronique", "maison"},
		tip:           "Créez une liste de souhaits Amazon à l'avance et activez les alertes de prix.",
	},
	{
		id: "back-to-school-2026", name: "Rentrée Scolaire", eventType: "back_to_school",
		start: "2026-08-15", end: "2026-09-15",
		discountRange: "10-30%",
		categories:    []string{"informatique", "fournitures"},
		tip:           "Les ordinateurs et tablettes sont à leur plus bas prix de l'année en août.",
	},
	{
		id: "black-friday-2026", name: "Black Friday", eventType: "flash",
		start: "2026-11-27", end: "2026-11-27",
		discountRange: "30-70%",
		categories:    []string{"électronique", "mode", "maison"},
		tip:           "Comparez les prix à l'avance : certains commerçants gonflent les prix avant de les baisser.",
	},
	{
		id: "cyber-monday-2026", name: "Cyber Monday", eventType: "flash",
		start: "2026-11-30", end: "2026-11-30",
		discountRange: "20-50%",
		categories:    []string{"électronique", "informatique"},
		tip:           "Idéal pour les achats tech de dernière minute avant Noël.",
	},
	{
		id: "noel-2026", name: "Noël", eventType: "fete",
		start: "2026-12-01", end: "2026-12-24",
		discountRange: "10-30%",
		categories:    []string{"jouets", "high-tech", "mode"},
		tip:           "Achetez avant le 15 décembre pour garantir la livraison avant les fêtes.",
	},
}

// Handler serves the sale calendar endpoint.
type Handler struct{}

// NewHandler returns a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// GetCalendar handles GET /api/sale-calendar.
// It computes daysUntil and isActive for each event relative to now,
// identifies the next upcoming and currently active events, and returns
// the full SaleCalendar response.
func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	events := make([]SaleEvent, 0, len(staticEvents))
	for _, raw := range staticEvents {
		start := parseDate(raw.start)
		end := parseDate(raw.end)
		// end is inclusive — treat the event as active through end-of-day
		endOfDay := end.Add(24*time.Hour - time.Second)

		isActive := !today.Before(start) && now.Before(endOfDay)
		daysUntil := int(start.Sub(today).Hours() / 24)

		events = append(events, SaleEvent{
			ID:            raw.id,
			Name:          raw.name,
			Type:          raw.eventType,
			StartDate:     start,
			EndDate:       end,
			DaysUntil:     daysUntil,
			IsActive:      isActive,
			DiscountRange: raw.discountRange,
			Categories:    raw.categories,
			Tip:           raw.tip,
		})
	}

	// Identify next upcoming and active events.
	var nextEvent *SaleEvent
	var activeEvent *SaleEvent
	for i := range events {
		e := &events[i]
		if e.IsActive && activeEvent == nil {
			activeEvent = e
		}
		if e.DaysUntil > 0 {
			if nextEvent == nil || e.DaysUntil < nextEvent.DaysUntil {
				nextEvent = e
			}
		}
	}

	cal := SaleCalendar{
		Events:      events,
		NextEvent:   nextEvent,
		ActiveEvent: activeEvent,
		GeneratedAt: now,
	}

	httputil.WriteJSON(w, http.StatusOK, cal)
}

// parseDate parses a "YYYY-MM-DD" string into a UTC midnight time.Time.
func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		// Should never happen with hard-coded dates; return zero value.
		return time.Time{}
	}
	return t
}
