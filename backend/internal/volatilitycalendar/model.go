package volatilitycalendar

// WeekData holds price-volatility data for a single ISO week.
type WeekData struct {
	Week      int     `json:"week"`
	Year      int     `json:"year"`
	Intensity float64 `json:"intensity"` // 0.0 (stable) – 1.0 (most volatile)
	MinPrice  float64 `json:"minPrice"`
	MaxPrice  float64 `json:"maxPrice"`
	Label     string  `json:"label"` // e.g. "Sem. 10"
}

// VolatilityCalendar is the response for GET /api/missions/{missionId}/items/{itemId}/volatility-calendar.
type VolatilityCalendar struct {
	ItemID           string     `json:"itemId"`
	Weeks            []WeekData `json:"weeks"`
	MostVolatileWeek *WeekData  `json:"mostVolatileWeek"`
	Recommendation   string     `json:"recommendation"`
}
