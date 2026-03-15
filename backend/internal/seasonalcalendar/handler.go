package seasonalcalendar

import (
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/httputil"
)

// cacheEntry holds a cached CategoryCalendar with expiration time
type cacheEntry struct {
	data      *CategoryCalendar
	expiresAt time.Time
}

// Handler serves the seasonal calendar endpoint.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // map[string]cacheEntry
	ttl   time.Duration
}

// NewHandler returns a new Handler.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
		ttl:  24 * time.Hour,
	}
}

// GetCalendar handles GET /api/missions/{missionId}/seasonal-calendar.
// It retrieves the mission category and returns a seasonal calendar with:
// - MonthRecommendations for each month (1-12) with scores and events
// - BestMonths (top 3) and WorstMonths (bottom 3)
// - CurrentMonthScore and CurrentAdvice
//
// Results are cached for 24 hours.
func (h *Handler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "missionId")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "Missing missionId")
		return
	}

	// Check cache
	if cached, ok := h.cache.Load(missionID); ok {
		entry := cached.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		// Cache expired, remove it
		h.cache.Delete(missionID)
	}

	// Query mission category
	var category string
	err := h.pool.QueryRow(r.Context(), "SELECT category FROM missions WHERE id = $1", missionID).Scan(&category)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "Mission not found")
		return
	}

	// Compute calendar
	cal := h.computeCalendar(category)

	// Cache it
	h.cache.Store(missionID, cacheEntry{
		data:      cal,
		expiresAt: time.Now().Add(h.ttl),
	})

	httputil.WriteJSON(w, http.StatusOK, cal)
}

// computeCalendar computes a CategoryCalendar for the given category.
func (h *Handler) computeCalendar(category string) *CategoryCalendar {
	now := time.Now()
	currentMonth := int(now.Month())

	// Base scores by month (before category boosts)
	baseScores := []int{
		50, // January
		40, // February
		45, // March
		50, // April
		55, // May
		60, // June
		55, // July
		50, // August
		55, // September
		60, // October
		80, // November (Black Friday)
		65, // December
	}

	// Category-specific multipliers/boosts
	boosts := h.getMonthBoosts(category)

	// Compute 12 MonthRecommendation entries
	recommendations := make([]MonthRecommendation, 12)
	for i := 0; i < 12; i++ {
		month := i + 1
		monthName := frenchMonthName(month)
		baseScore := baseScores[i]
		boost := boosts[month]
		score := baseScore + boost
		if score > 100 {
			score = 100
		}
		if score < 0 {
			score = 0
		}

		events := h.getMonthEvents(month, category)
		advice := h.getMonthAdvice(month, score)

		recommendations[i] = MonthRecommendation{
			Month:     month,
			MonthName: monthName,
			Score:     score,
			Events:    events,
			Advice:    advice,
		}
	}

	// Find best months (top 3) and worst months (bottom 3)
	sorted := make([]MonthRecommendation, len(recommendations))
	copy(sorted, recommendations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	bestMonths := make([]int, 0, 3)
	for i := 0; i < 3 && i < len(sorted); i++ {
		bestMonths = append(bestMonths, sorted[i].Month)
	}

	worstMonths := make([]int, 0, 3)
	for i := len(sorted) - 1; i >= len(sorted)-3 && i >= 0; i-- {
		worstMonths = append(worstMonths, sorted[i].Month)
	}
	// Reverse to maintain month order
	reverse(worstMonths)

	// Current month score and advice
	currentMonthScore := recommendations[currentMonth-1].Score
	currentMonthAdvice := h.getAdvice(currentMonthScore)

	return &CategoryCalendar{
		Category:             category,
		BestMonths:           bestMonths,
		WorstMonths:          worstMonths,
		MonthRecommendations: recommendations,
		CurrentMonthScore:    currentMonthScore,
		CurrentAdvice:        currentMonthAdvice,
	}
}

// getMonthBoosts returns category-specific score boosts for each month.
// Boosts are applied to base scores; they can be negative.
func (h *Handler) getMonthBoosts(category string) map[int]int {
	boosts := make(map[int]int)

	// Set all to 0 by default
	for i := 1; i <= 12; i++ {
		boosts[i] = 0
	}

	// Normalize category to lowercase for comparison
	cat := strings.ToLower(strings.TrimSpace(category))

	// Generic events (apply to most categories)
	boosts[1] += 20  // Soldes d'hiver
	boosts[2] += 10  // Soldes d'hiver fin
	boosts[11] += 50 // Black Friday, Cyber Monday
	boosts[12] += 10 // Christmas deals

	// Category-specific boosts
	if isCategory(cat, "électronique", "tech", "informatique", "high-tech", "téléphone", "laptop", "ordinateur", "écran") {
		boosts[1] += 20  // Soldes d'hiver
		boosts[9] += 20  // Rentrée (back to school)
		boosts[11] += 20 // Black Friday (extra for electronics)
		boosts[12] -= 10 // Avoid pre-Christmas rush
	}

	if isCategory(cat, "mode", "vêtements", "chaussures", "textile", "habit", "robe", "pantalon", "tee-shirt") {
		boosts[1] += 20  // Soldes d'hiver
		boosts[2] += 10  // Soldes d'hiver fin
		boosts[3] += 15  // Spring arrivals starting
		boosts[6] += 20  // Soldes d'été start
		boosts[7] += 15  // Soldes d'été
		boosts[11] += 15 // Black Friday fashion
	}

	if isCategory(cat, "maison", "meubles", "décoration", "cuisine", "électroménager", "four", "réfrigérateur", "lave-linge") {
		boosts[1] += 15   // Soldes d'hiver
		boosts[3] += 10   // Spring home refresh
		boosts[6] += 10   // Summer refresh
		boosts[11] += 20  // Black Friday
	}

	if isCategory(cat, "jardin", "jardinage", "plantes", "outils", "piscine", "mobilier de jardin") {
		boosts[3] += 20   // Spring gardening season
		boosts[4] += 10   // April gardening
		boosts[5] += 5    // May
		boosts[6] += 5    // June
	}

	if isCategory(cat, "sport", "fitness", "équipement sportif", "vélo", "chaussures de sport", "accessoires sport") {
		boosts[1] += 15   // Soldes d'hiver (winter sports)
		boosts[3] += 10   // Spring sports start
		boosts[6] += 15   // Summer sports
		boosts[7] += 10   // Soldes d'été
		boosts[11] += 10  // Black Friday
	}

	if isCategory(cat, "beauté", "cosmétique", "parfum", "soin", "maquillage", "shampooing") {
		boosts[2] += 20   // Valentine's Day (early beauty)
		boosts[5] += 10   // Mother's Day (May)
		boosts[11] += 15  // Black Friday
	}

	if isCategory(cat, "livres", "littérature", "papeterie", "fournitures scolaires", "école") {
		boosts[8] += 20   // Back to school (August)
		boosts[9] += 20   // Rentrée (September)
	}

	if isCategory(cat, "jouets", "jeux") {
		boosts[12] += 30  // Christmas (major toy season)
		boosts[11] += 20  // Black Friday
	}

	if isCategory(cat, "bijoux", "montres", "accessoires") {
		boosts[2] += 20   // Valentine's Day
		boosts[5] += 15   // Mother's Day
		boosts[12] += 15  // Christmas gifts
	}

	return boosts
}

// getMonthEvents returns a list of events for the given month and category.
func (h *Handler) getMonthEvents(month int, category string) []string {
	var events []string

	cat := strings.ToLower(strings.TrimSpace(category))

	switch month {
	case 1:
		events = append(events, "Soldes d'hiver")
		if isCategory(cat, "électronique", "tech") {
			events = append(events, "Promotions tech")
		}
	case 2:
		if isCategory(cat, "beauté", "cosmétique", "bijoux") {
			events = append(events, "Saint-Valentin")
		}
		events = append(events, "Fin soldes d'hiver")
	case 3:
		if isCategory(cat, "jardin", "jardinage") {
			events = append(events, "Printemps (saison jardinage)")
		}
	case 4:
	case 5:
		if isCategory(cat, "beauté", "cosmétique", "bijoux") {
			events = append(events, "Fête des Mères")
		}
	case 6:
		events = append(events, "Soldes d'été")
	case 7:
		events = append(events, "Soldes d'été")
	case 8:
		if isCategory(cat, "livres", "école", "fournitures") {
			events = append(events, "Rentrée scolaire (début)")
		}
	case 9:
		if isCategory(cat, "livres", "école", "informatique") {
			events = append(events, "Rentrée (back to school)")
		}
	case 10:
	case 11:
		events = append(events, "Black Friday")
		events = append(events, "Cyber Monday")
	case 12:
		if isCategory(cat, "jouets", "jeux") {
			events = append(events, "Noël (saison majeure)")
		} else {
			events = append(events, "Noël")
		}
	}

	return events
}

// getMonthAdvice returns French advice based on the month's score.
func (h *Handler) getMonthAdvice(month int, score int) string {
	switch month {
	case 1:
		return "Soldes d'hiver — les prix baissent drastiquement, surtout les dernières semaines"
	case 2:
		return "Fin des soldes d'hiver ; bons prix avant les arrivages de printemps"
	case 3:
		return "Arrivages de printemps ; prix stables, mais stock varié"
	case 4:
		return "Prix stables ; attendez mai/juin pour les vraies promotions"
	case 5:
		return "Prix avant-vente été ; attendez juin pour les soldes"
	case 6:
		return "Début soldes d'été — premières 48h offrent les meilleures remises"
	case 7:
		return "Soldes d'été en cours ; bon stock et remises substantielles"
	case 8:
		return "Rentrée ; prix réalistes, stock neuf pour la rentrée"
	case 9:
		return "Post-rentrée ; prix normalisés mais risque de ruptures"
	case 10:
		return "Avant Black Friday ; attendez novembre pour les meilleures affaires"
	case 11:
		return "Black Friday & Cyber Monday — LES meilleures affaires de l'année"
	case 12:
		return "Noël ; promotions tardives mais risque d'épuisement de stock"
	default:
		return ""
	}
}

// getAdvice returns a general French advice message based on the score.
func (h *Handler) getAdvice(score int) string {
	if score > 70 {
		return "Excellent moment pour acheter — profitez des bonnes affaires !"
	}
	if score >= 50 {
		return "Bon moment pour acheter — prix raisonnables"
	}
	return "Attendez les prochaines promotions pour de meilleures affaires"
}

// frenchMonthName returns the French name of the month.
func frenchMonthName(month int) string {
	names := []string{
		"Janvier", "Février", "Mars", "Avril", "Mai", "Juin",
		"Juillet", "Août", "Septembre", "Octobre", "Novembre", "Décembre",
	}
	if month < 1 || month > 12 {
		return ""
	}
	return names[month-1]
}

// isCategory checks if the target category matches any of the provided category keywords.
func isCategory(target string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(target, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// reverse reverses a slice of integers in place.
func reverse(s []int) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
