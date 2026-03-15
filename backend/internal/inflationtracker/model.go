package inflationtracker

// InflationItem holds the computed inflation impact for a single shopping item.
type InflationItem struct {
	ItemID              string  `json:"itemId"`
	Name                string  `json:"name"`
	Price               float64 `json:"price"`
	AnnualInflationRate float64 `json:"annualInflationRate"`
	InflationPer30Days  float64 `json:"inflationPer30Days"`
	WaitCost3Months     float64 `json:"waitCost3Months"`
	Recommendation      string  `json:"recommendation"`
}

// InflationImpact is the full response for GET /api/missions/{missionId}/inflation-impact.
type InflationImpact struct {
	MissionID            string          `json:"missionId"`
	Items                []InflationItem `json:"items"`
	TotalWaitCost3Months float64         `json:"totalWaitCost3Months"`
	AvgInflationRate     float64         `json:"avgInflationRate"`
}

// inflationRates maps French category prefixes (lower-case) to annual CPI rates.
var inflationRates = map[string]float64{
	"électronique":    0.02,
	"électroménager":  0.03,
	"alimentation":    0.04,
	"habillement":     0.025,
	"mobilier":        0.035,
	"automobile":      0.03,
	"loisirs":         0.02,
	"sport":           0.025,
	"informatique":    -0.01,
	"téléphonie":      -0.02,
}

const defaultRate = 0.025

// rateForCategory returns the annual inflation rate for a given category string.
// Matching is case-insensitive prefix-based.
func rateForCategory(cat string) float64 {
	if cat == "" {
		return defaultRate
	}

	// Normalise: lowercase, fold accented characters to a comparable form.
	// We compare the stored key directly against a lower-case version of cat.
	lower := toLower(cat)
	for key, rate := range inflationRates {
		if hasPrefix(lower, key) {
			return rate
		}
	}
	return defaultRate
}

// toLower is a simple ASCII + common accented-char lowercaser sufficient for
// the French category names used here.
func toLower(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if r >= 'A' && r <= 'Z' {
			runes[i] = r + ('a' - 'A')
		}
		// Handle common upper-case French accented letters.
		switch r {
		case 'É', 'È', 'Ê', 'Ë':
			runes[i] = 'é'
		case 'À', 'Â':
			runes[i] = 'à'
		case 'Î', 'Ï':
			runes[i] = 'î'
		case 'Ô':
			runes[i] = 'ô'
		case 'Ù', 'Û', 'Ü':
			runes[i] = 'ù'
		case 'Ç':
			runes[i] = 'ç'
		}
	}
	return string(runes)
}

// hasPrefix returns true when s starts with prefix (both already lower-cased).
func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

// recommendationFor returns a French recommendation based on the annual rate.
func recommendationFor(rate float64) string {
	switch {
	case rate < 0:
		return "Ce type de produit baisse en prix, attendez !"
	case rate > 0.03:
		return "Inflation forte dans cette catégorie — achetez vite"
	default:
		return "Inflation modérée — suivez le prix"
	}
}
