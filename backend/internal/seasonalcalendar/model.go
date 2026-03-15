package seasonalcalendar

type MonthRecommendation struct {
	Month       int      `json:"month"`       // 1-12
	MonthName   string   `json:"monthName"`   // French month name
	Score       int      `json:"score"`       // 0-100 (how good is this month to buy)
	Events      []string `json:"events"`      // e.g. ["Soldes d'hiver", "Black Friday"]
	Advice      string   `json:"advice"`      // French advice
}

type CategoryCalendar struct {
	Category             string                `json:"category"`
	BestMonths           []int                 `json:"bestMonths"`                // top 3 month numbers
	WorstMonths          []int                 `json:"worstMonths"`               // worst 3 month numbers
	MonthRecommendations []MonthRecommendation `json:"monthRecommendations"`
	CurrentMonthScore    int                   `json:"currentMonthScore"`
	CurrentAdvice        string                `json:"currentAdvice"`
}
