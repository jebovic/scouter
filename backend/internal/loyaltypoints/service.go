package loyaltypoints

import (
	"errors"
	"strings"
)

// ErrProgramNotFound is returned when no program matches the given name/retailer.
var ErrProgramNotFound = errors.New("loyalty program not found")

// registry is the hardcoded list of French loyalty programs.
var registry = []LoyaltyProgram{
	{
		Name:          "Carrefour PASS",
		Retailer:      "Carrefour",
		PointsRate:    0.5,
		CashbackRate:  0,
		MinRedemption: 10,
		Notes:         "1 pt = 0.01€ — cumulable en caisse ou en ligne",
	},
	{
		Name:          "Leclerc Avantage",
		Retailer:      "Leclerc",
		PointsRate:    0,
		CashbackRate:  0.5,
		MinRedemption: 5,
		Notes:         "0.5 % remboursé sur ticket caisse, rechargé chaque trimestre",
	},
	{
		Name:          "Fnac+",
		Retailer:      "Fnac",
		PointsRate:    1,
		CashbackRate:  0,
		MinRedemption: 5,
		Notes:         "1 pt = 0.01€ — adhésion payante (49.99€/an)",
	},
	{
		Name:          "Amazon Prime Rewards",
		Retailer:      "Amazon",
		PointsRate:    0,
		CashbackRate:  1,
		MinRedemption: 0,
		Notes:         "1 % cashback sur Amazon.fr pour les membres Prime (taux variable selon offres)",
	},
	{
		Name:          "Rakuten Club",
		Retailer:      "Rakuten",
		PointsRate:    0,
		CashbackRate:  5,
		MinRedemption: 5,
		Notes:         "Jusqu'à 5 % de Super Points remboursés — taux réel variable par marchand",
	},
	{
		Name:          "Cdiscount à volonté",
		Retailer:      "Cdiscount",
		PointsRate:    0,
		CashbackRate:  1,
		MinRedemption: 5,
		Notes:         "1 % cashback sur les achats éligibles, adhésion 29€/an",
	},
	{
		Name:          "Darty Max",
		Retailer:      "Darty",
		PointsRate:    0,
		CashbackRate:  0,
		MinRedemption: 0,
		Notes:         "Abonnement mensuel (9.99€/mois) — pas de points, réparations illimitées incluses",
	},
	{
		Name:          "La Redoute Fidélité",
		Retailer:      "La Redoute",
		PointsRate:    1,
		CashbackRate:  0,
		MinRedemption: 10,
		Notes:         "1 pt par euro dépensé, bons de réduction émis par palier de 100 pts (= 5€)",
	},
	{
		Name:          "Boulanger Club",
		Retailer:      "Boulanger",
		PointsRate:    0.5,
		CashbackRate:  0,
		MinRedemption: 10,
		Notes:         "0.5 pt/€ — 100 pts = 5€ de bon d'achat",
	},
	{
		Name:          "Monoprix M comme Miles",
		Retailer:      "Monoprix",
		PointsRate:    1,
		CashbackRate:  0,
		MinRedemption: 10,
		Notes:         "1 M comme Mile par euro dépensé — bons de réduction à partir de 100 points (= 5€)",
	},
}

// Service provides access to the loyalty-program registry.
type Service struct{}

// NewService returns a new Service.
func NewService() *Service { return &Service{} }

// ListPrograms returns all known loyalty programs.
func (s *Service) ListPrograms() []LoyaltyProgram {
	out := make([]LoyaltyProgram, len(registry))
	copy(out, registry)
	return out
}

// Calculate computes the points earned and estimated cashback for a purchase.
// It looks up the program by programName first, falling back to retailer match.
func (s *Service) Calculate(req CalculateRequest) (CalculateResponse, error) {
	prog, err := s.findProgram(req.Retailer, req.ProgramName)
	if err != nil {
		return CalculateResponse{}, err
	}

	pointsEarned := prog.PointsRate * req.Price
	// Each point is worth 0.01€ by convention unless cashback rate applies.
	var estimatedCashback float64
	if prog.CashbackRate > 0 {
		estimatedCashback = (prog.CashbackRate / 100) * req.Price
	} else {
		// points: 1 pt = 0.01€
		estimatedCashback = pointsEarned * 0.01
	}

	return CalculateResponse{
		PointsEarned:      pointsEarned,
		EstimatedCashback: estimatedCashback,
		Notes:             prog.Notes,
	}, nil
}

// findProgram locates a program by exact name match first, then retailer prefix.
func (s *Service) findProgram(retailer, programName string) (LoyaltyProgram, error) {
	nameLower := strings.ToLower(strings.TrimSpace(programName))
	retailerLower := strings.ToLower(strings.TrimSpace(retailer))

	// Exact name match
	if nameLower != "" {
		for _, p := range registry {
			if strings.ToLower(p.Name) == nameLower {
				return p, nil
			}
		}
	}

	// Fallback: retailer prefix match
	if retailerLower != "" {
		for _, p := range registry {
			if strings.HasPrefix(strings.ToLower(p.Retailer), retailerLower) {
				return p, nil
			}
		}
	}

	return LoyaltyProgram{}, ErrProgramNotFound
}
