package health

// ComputeSignals computes the four mission health signal scores using pure Go
// logic (no LLM required). Weights must sum to 1.0.
//
// Parameters:
//   - optionCount: total number of options researched
//   - researchCount: unused reserved parameter (future use)
//   - budgetUsedPct: percentage of budget allocated (> 0 means budget is set)
//   - phase: mission phase — "researching" | "comparing" | "buying" | "done"
//   - itemCount: number of shopping items (reserved for future use)
func ComputeSignals(optionCount, researchCount int, budgetUsedPct float64, phase string, itemCount int) []SignalScore {
	return []SignalScore{
		computeResearchDepth(optionCount),
		computeOptionCoverage(optionCount),
		computeBudgetClarity(budgetUsedPct),
		computeDecisionProgress(phase),
	}
}

// ComputeOverallScore computes the weighted average of the signal scores,
// rounded to the nearest integer.
func ComputeOverallScore(signals []SignalScore) int {
	var weighted float64
	for _, s := range signals {
		weighted += float64(s.Score) * s.Weight
	}
	return int(weighted)
}

// ComputeGrade converts an overall score (0–100) to a letter grade.
//   - 90–100 → A
//   - 80–89  → B
//   - 70–79  → C
//   - 60–69  → D
//   - 0–59   → F
func ComputeGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// ── signal constructors ───────────────────────────────────────────────────────

func computeResearchDepth(optionCount int) SignalScore {
	score := min100(optionCount * 20)
	return SignalScore{
		Name:        "Research Depth",
		Score:       score,
		Weight:      0.25,
		Description: "Nombre d'options recherchées pour cette mission",
		Icon:        "🔬",
	}
}

func computeOptionCoverage(optionCount int) SignalScore {
	score := min100(optionCount * 20)
	return SignalScore{
		Name:        "Option Coverage",
		Score:       score,
		Weight:      0.25,
		Description: "Couverture des options disponibles sur le marché",
		Icon:        "📊",
	}
}

func computeBudgetClarity(budgetUsedPct float64) SignalScore {
	score := 30
	if budgetUsedPct > 0 {
		score = 100
	}
	return SignalScore{
		Name:        "Budget Clarity",
		Score:       score,
		Weight:      0.20,
		Description: "Clarté de l'utilisation du budget alloué",
		Icon:        "💰",
	}
}

func computeDecisionProgress(phase string) SignalScore {
	phaseScores := map[string]int{
		"researching": 20,
		"comparing":   50,
		"buying":      80,
		"done":        100,
	}
	score, ok := phaseScores[phase]
	if !ok {
		score = 20
	}
	return SignalScore{
		Name:        "Decision Progress",
		Score:       score,
		Weight:      0.30,
		Description: "Avancement dans le processus de décision d'achat",
		Icon:        "🎯",
	}
}

// min100 clamps v to [0, 100].
func min100(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
