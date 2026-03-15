package dealcalendar

import "time"

// Calendar is a static registry of French market sales events for 2026.
type Calendar struct {
	events []DealEvent
}

// NewCalendar creates a new calendar with hardcoded French sales events.
func NewCalendar() *Calendar {
	events := []DealEvent{
		{
			Name:          "Soldes d'hiver",
			StartDate:     time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 2, 18, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "all",
			DiscountLevel: "high",
			Description:   "Les soldes d'hiver offrent les plus grandes réductions de l'année sur l'habillement, chaussures et accessoires.",
		},
		{
			Name:          "French Days printemps",
			StartDate:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 5, 4, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "all",
			DiscountLevel: "high",
			Description:   "French Days est l'événement de shopping annuel français avec réductions massives sur tous les secteurs.",
		},
		{
			Name:          "Fête des mères promo",
			StartDate:     time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 5, 25, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "all",
			DiscountLevel: "medium",
			Description:   "Promotions spéciales pour la Fête des mères avec des offres réduites sur les cadeaux et produits de luxe.",
		},
		{
			Name:          "Amazon Prime Day",
			StartDate:     time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 7, 9, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "electronics",
			DiscountLevel: "high",
			Description:   "Amazon Prime Day offre des réductions exceptionnelles sur l'électronique, livres, maison et bien plus.",
		},
		{
			Name:          "Soldes d'été",
			StartDate:     time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 8, 5, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "all",
			DiscountLevel: "high",
			Description:   "Les soldes d'été marquent le début des plus grandes réductions estivales sur la mode et les accessoires.",
		},
		{
			Name:          "La Rentrée tech",
			StartDate:     time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 9, 5, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "electronics",
			DiscountLevel: "medium",
			Description:   "La Rentrée tech propose des réductions attractives sur les appareils électroniques et l'informatique.",
		},
		{
			Name:          "Darty Days",
			StartDate:     time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 10, 5, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "electronics",
			DiscountLevel: "medium",
			Description:   "Darty Days offre des prix compétitifs sur l'électroménager, l'audio-vidéo et l'informatique.",
		},
		{
			Name:          "Fnac Days",
			StartDate:     time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 10, 12, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "electronics,computing",
			DiscountLevel: "medium",
			Description:   "Fnac Days présente des offres réduites sur les livres, jeux vidéo, électronique et informatique.",
		},
		{
			Name:          "Black Friday",
			StartDate:     time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 11, 27, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "all",
			DiscountLevel: "high",
			Description:   "Black Friday propose les réductions les plus agressives de l'année sur tous les secteurs.",
		},
		{
			Name:          "Cyber Monday",
			StartDate:     time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 11, 30, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "electronics,computing",
			DiscountLevel: "high",
			Description:   "Cyber Monday offre des réductions massives sur l'électronique et l'informatique en ligne.",
		},
		{
			Name:          "Noël promotions",
			StartDate:     time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			EndDate:       time.Date(2026, 12, 24, 23, 59, 59, 0, time.UTC),
			CategoryHint:  "all",
			DiscountLevel: "medium",
			Description:   "Les promotions de Noël offrent des réductions attrayantes sur les cadeaux et produits populaires.",
		},
	}
	return &Calendar{events: events}
}

// GetUpcoming returns all events that start on or after today, sorted by start date.
func (c *Calendar) GetUpcoming(now time.Time) []DealEvent {
	var result []DealEvent
	for _, event := range c.events {
		if event.StartDate.After(now) || event.StartDate.Equal(now) {
			result = append(result, event)
		}
	}
	return result
}

// FilterByCategory filters events by category hint. Returns all events if category is empty.
// An event matches if its categoryHint is "all" or contains the category.
func (c *Calendar) FilterByCategory(events []DealEvent, category string) []DealEvent {
	if category == "" {
		return events
	}
	var result []DealEvent
	for _, event := range events {
		if event.CategoryHint == "all" || contains(event.CategoryHint, category) {
			result = append(result, event)
		}
	}
	return result
}

// GetAll returns all events sorted by start date.
func (c *Calendar) GetAll() []DealEvent {
	return c.events
}

// contains checks if a comma-separated string contains a value.
func contains(csv, value string) bool {
	// Simple split and check for comma-separated values
	fields := split(csv, ",")
	for _, f := range fields {
		if f == value {
			return true
		}
	}
	return false
}

// split is a simple string split by delimiter, trimming whitespace.
func split(s, sep string) []string {
	var result []string
	var current string
	for _, ch := range s {
		if string(ch) == sep {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
