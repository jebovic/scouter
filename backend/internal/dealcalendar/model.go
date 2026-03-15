package dealcalendar

import "time"

// DealEvent represents a French seasonal sales event.
type DealEvent struct {
	Name          string    `json:"name"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate"`
	CategoryHint  string    `json:"categoryHint"` // "all" | "electronics" | "fashion" | "travel" | "computing"
	DiscountLevel string    `json:"discountLevel"` // "low" | "medium" | "high"
	Description   string    `json:"description"`
}
