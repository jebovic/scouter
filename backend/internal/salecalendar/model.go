package salecalendar

import "time"

// SaleEvent represents a French seasonal or promotional shopping event.
type SaleEvent struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`          // "soldes", "flash", "fete", "back_to_school"
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	DaysUntil     int       `json:"daysUntil"`     // negative if past
	IsActive      bool      `json:"isActive"`
	DiscountRange string    `json:"discountRange"` // e.g. "20-70%"
	Categories    []string  `json:"categories"`    // best categories to buy during this event
	Tip           string    `json:"tip"`
}

// SaleCalendar is the top-level response for GET /api/sale-calendar.
type SaleCalendar struct {
	Events      []SaleEvent `json:"events"`
	NextEvent   *SaleEvent  `json:"nextEvent,omitempty"`
	ActiveEvent *SaleEvent  `json:"activeEvent,omitempty"`
	GeneratedAt time.Time   `json:"generatedAt"`
}
