package holidayalert

import "time"

// EventType classifies a calendar entry.
type EventType string

const (
	EventTypeHoliday      EventType = "holiday"
	EventTypeShoppingEvent EventType = "shopping_event"
)

// HolidayEvent is a single entry in the French public holiday / shopping event
// calendar for 2026.
type HolidayEvent struct {
	Name      string    `json:"name"`
	Date      string    `json:"date"`               // "YYYY-MM-DD"
	EndDate   string    `json:"endDate,omitempty"`  // non-empty for multi-day events
	Type      EventType `json:"type"`
	DaysUntil int       `json:"daysUntil"`
	IsActive  bool      `json:"isActive"`
	Advice    string    `json:"advice"`
}

// HolidayAlertResponse is the top-level response for GET /api/holidays-and-events.
type HolidayAlertResponse struct {
	Events            []HolidayEvent `json:"events"`
	NextShoppingEvent *HolidayEvent  `json:"nextShoppingEvent,omitempty"`
	ActiveEvent       *HolidayEvent  `json:"activeEvent,omitempty"`
	GeneratedAt       time.Time      `json:"generatedAt"`
}
