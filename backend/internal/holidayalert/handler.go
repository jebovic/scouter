package holidayalert

import (
	"net/http"
	"time"

	"github.com/jibei/scouter/internal/httputil"
)

// rawEvent is the compile-time dataset entry before date arithmetic.
type rawEvent struct {
	name      string
	date      string // "YYYY-MM-DD", start date
	endDate   string // non-empty for multi-day events
	eventType EventType
	advice    string
}

// staticEvents is the hard-coded 2026 French holiday + shopping event dataset.
// Today is assumed to be 2026-03-15 for daysUntil / isActive computation, but
// the handler uses time.Now() so it stays correct in production.
var staticEvents = []rawEvent{
	// ── Public holidays ────────────────────────────────────────────────────────
	{
		name:      "Jour de l'An",
		date:      "2026-01-01",
		eventType: EventTypeHoliday,
		advice:    "Les magasins sont souvent fermés — planifiez vos achats à l'avance.",
	},
	{
		name:      "Lundi de Pâques",
		date:      "2026-04-06",
		eventType: EventTypeHoliday,
		advice:    "Profitez du weekend pascal pour comparer les offres en ligne.",
	},
	{
		name:      "Fête du Travail",
		date:      "2026-05-01",
		eventType: EventTypeHoliday,
		advice:    "La plupart des commerces sont fermés — commandez en ligne à la place.",
	},
	{
		name:      "Victoire 1945",
		date:      "2026-05-08",
		eventType: EventTypeHoliday,
		advice:    "Bon moment pour des achats en ligne pendant le pont.",
	},
	{
		name:      "Ascension",
		date:      "2026-05-14",
		eventType: EventTypeHoliday,
		advice:    "Pont prolongé — idéal pour repérer les offres en ligne.",
	},
	{
		name:      "Lundi de Pentecôte",
		date:      "2026-05-25",
		eventType: EventTypeHoliday,
		advice:    "Les ventes flash se multiplient avant ce férié.",
	},
	{
		name:      "Fête Nationale",
		date:      "2026-07-14",
		eventType: EventTypeHoliday,
		advice:    "Certains commerçants proposent des offres patriotiques — vérifiez vos alertes.",
	},
	{
		name:      "Assomption",
		date:      "2026-08-15",
		eventType: EventTypeHoliday,
		advice:    "En plein milieu des soldes d'été — profitez-en avant la fermeture estivale.",
	},
	{
		name:      "Toussaint",
		date:      "2026-11-01",
		eventType: EventTypeHoliday,
		advice:    "À quelques semaines du Black Friday — surveillez les précommandes.",
	},
	{
		name:      "Armistice",
		date:      "2026-11-11",
		eventType: EventTypeHoliday,
		advice:    "Pont stratégique : les pré-ventes Black Friday débutent souvent ici.",
	},
	{
		name:      "Noël",
		date:      "2026-12-25",
		eventType: EventTypeHoliday,
		advice:    "Les commerces sont fermés — assurez-vous d'avoir commandé avant le 20 décembre.",
	},

	// ── Shopping events ────────────────────────────────────────────────────────
	{
		name:      "Soldes d'hiver",
		date:      "2026-01-08",
		endDate:   "2026-02-18",
		eventType: EventTypeShoppingEvent,
		advice:    "Profitez des soldes d'hiver pour acheter maintenant — les remises atteignent 70 %.",
	},
	{
		name:      "French Days printemps",
		date:      "2026-03-06",
		endDate:   "2026-03-11",
		eventType: EventTypeShoppingEvent,
		advice:    "Les French Days de printemps approchent — préparez votre liste de souhaits.",
	},
	{
		name:      "Soldes d'été",
		date:      "2026-06-25",
		endDate:   "2026-08-05",
		eventType: EventTypeShoppingEvent,
		advice:    "Les premières 48h des soldes d'été offrent les meilleures remises.",
	},
	{
		name:      "Black Friday",
		date:      "2026-11-27",
		eventType: EventTypeShoppingEvent,
		advice:    "Comparez les prix à l'avance pour détecter les fausses promotions.",
	},
	{
		name:      "Cyber Monday",
		date:      "2026-11-30",
		eventType: EventTypeShoppingEvent,
		advice:    "Idéal pour les achats tech de dernière minute avant Noël.",
	},
	{
		name:      "French Days spécial Noël",
		date:      "2026-12-05",
		endDate:   "2026-12-10",
		eventType: EventTypeShoppingEvent,
		advice:    "Dernière chance pour des remises significatives avant les fêtes.",
	},
}

// Handler serves GET /api/holidays-and-events.
type Handler struct{}

// NewHandler returns a new Handler.
func NewHandler() *Handler {
	return &Handler{}
}

// List handles GET /api/holidays-and-events.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	events := make([]HolidayEvent, 0, len(staticEvents))
	for _, raw := range staticEvents {
		start := parseDate(raw.date)
		isActive := false
		daysUntil := int(start.Sub(today).Hours() / 24)

		if raw.endDate != "" {
			end := parseDate(raw.endDate)
			// Event is active from start (inclusive) through end (inclusive, end-of-day).
			endOfDay := end.Add(24*time.Hour - time.Second)
			isActive = !today.Before(start) && now.Before(endOfDay)
		} else {
			// Single-day event: active only on the event day itself.
			isActive = today.Equal(start)
		}

		events = append(events, HolidayEvent{
			Name:      raw.name,
			Date:      raw.date,
			EndDate:   raw.endDate,
			Type:      raw.eventType,
			DaysUntil: daysUntil,
			IsActive:  isActive,
			Advice:    raw.advice,
		})
	}

	// Identify the next upcoming shopping event and any active event.
	var nextShoppingEvent *HolidayEvent
	var activeEvent *HolidayEvent

	for i := range events {
		e := &events[i]
		if e.IsActive && activeEvent == nil {
			activeEvent = e
		}
		if e.Type == EventTypeShoppingEvent && e.DaysUntil > 0 {
			if nextShoppingEvent == nil || e.DaysUntil < nextShoppingEvent.DaysUntil {
				nextShoppingEvent = e
			}
		}
	}

	resp := HolidayAlertResponse{
		Events:            events,
		NextShoppingEvent: nextShoppingEvent,
		ActiveEvent:       activeEvent,
		GeneratedAt:       now,
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

// parseDate parses a "YYYY-MM-DD" string into a UTC midnight time.Time.
func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
