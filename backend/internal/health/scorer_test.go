package health_test

import (
	"testing"

	"github.com/jibei/scouter/internal/health"
)

// TestComputeSignals_ResearchDepth verifies Research Depth signal scoring.
func TestComputeSignals_ResearchDepth(t *testing.T) {
	// 0 options → research score = 0
	signals := health.ComputeSignals(0, 0, 0, "researching", 0)
	researchSig := findSignal(t, signals, "Research Depth")
	if researchSig.Score != 0 {
		t.Errorf("expected Research Depth score 0 for 0 options, got %d", researchSig.Score)
	}

	// 3 options → score = min(100, 3*20) = 60
	signals = health.ComputeSignals(3, 0, 0, "researching", 0)
	researchSig = findSignal(t, signals, "Research Depth")
	if researchSig.Score != 60 {
		t.Errorf("expected Research Depth score 60 for 3 options, got %d", researchSig.Score)
	}

	// 5+ options → score = 100 (capped)
	signals = health.ComputeSignals(6, 0, 0, "researching", 0)
	researchSig = findSignal(t, signals, "Research Depth")
	if researchSig.Score != 100 {
		t.Errorf("expected Research Depth score 100 (capped) for 6 options, got %d", researchSig.Score)
	}
}

// TestComputeSignals_OptionCoverage verifies Option Coverage signal scoring.
func TestComputeSignals_OptionCoverage(t *testing.T) {
	// 0 options → option coverage = 0
	signals := health.ComputeSignals(0, 0, 0, "researching", 0)
	sig := findSignal(t, signals, "Option Coverage")
	if sig.Score != 0 {
		t.Errorf("expected Option Coverage score 0 for 0 options, got %d", sig.Score)
	}

	// 5 options → score = 100
	signals = health.ComputeSignals(5, 0, 0, "researching", 0)
	sig = findSignal(t, signals, "Option Coverage")
	if sig.Score != 100 {
		t.Errorf("expected Option Coverage score 100 for 5 options, got %d", sig.Score)
	}

	// 2 options → score = min(100, 2*20) = 40
	signals = health.ComputeSignals(2, 0, 0, "researching", 0)
	sig = findSignal(t, signals, "Option Coverage")
	if sig.Score != 40 {
		t.Errorf("expected Option Coverage score 40 for 2 options, got %d", sig.Score)
	}
}

// TestComputeSignals_BudgetClarity verifies Budget Clarity signal scoring.
func TestComputeSignals_BudgetClarity(t *testing.T) {
	// Budget used > 0 → score = 100
	signals := health.ComputeSignals(0, 0, 50.0, "researching", 0)
	sig := findSignal(t, signals, "Budget Clarity")
	if sig.Score != 100 {
		t.Errorf("expected Budget Clarity score 100 when budgetUsedPct > 0, got %d", sig.Score)
	}

	// Budget used = 0 → score = 30
	signals = health.ComputeSignals(0, 0, 0, "researching", 0)
	sig = findSignal(t, signals, "Budget Clarity")
	if sig.Score != 30 {
		t.Errorf("expected Budget Clarity score 30 when budgetUsedPct = 0, got %d", sig.Score)
	}
}

// TestComputeSignals_DecisionProgress verifies Decision Progress signal scoring for each phase.
func TestComputeSignals_DecisionProgress(t *testing.T) {
	cases := []struct {
		phase    string
		expected int
	}{
		{"researching", 20},
		{"comparing", 50},
		{"buying", 80},
		{"done", 100},
		{"unknown", 20}, // unknown phase → default to 20
	}

	for _, tc := range cases {
		t.Run(tc.phase, func(t *testing.T) {
			signals := health.ComputeSignals(0, 0, 0, tc.phase, 0)
			sig := findSignal(t, signals, "Decision Progress")
			if sig.Score != tc.expected {
				t.Errorf("phase=%s: expected Decision Progress score %d, got %d", tc.phase, tc.expected, sig.Score)
			}
		})
	}
}

// TestComputeSignals_SignalCount verifies exactly 4 signals are returned.
func TestComputeSignals_SignalCount(t *testing.T) {
	signals := health.ComputeSignals(2, 1, 50.0, "comparing", 3)
	if len(signals) != 4 {
		t.Errorf("expected 4 signals, got %d", len(signals))
	}
}

// TestComputeSignals_WeightsSum verifies total weights sum to 1.0.
func TestComputeSignals_WeightsSum(t *testing.T) {
	signals := health.ComputeSignals(3, 2, 75.0, "comparing", 5)
	var total float64
	for _, s := range signals {
		total += s.Weight
	}
	// Allow small floating point tolerance
	if total < 0.99 || total > 1.01 {
		t.Errorf("expected weights to sum to 1.0, got %.4f", total)
	}
}

// TestComputeOverallScore verifies weighted average calculation.
func TestComputeOverallScore(t *testing.T) {
	// All signals at 100 → overall = 100
	signals := health.ComputeSignals(6, 3, 80.0, "done", 5)
	overall := health.ComputeOverallScore(signals)
	if overall != 100 {
		t.Errorf("expected overall 100 when all signals max, got %d", overall)
	}

	// All signals at 0 → overall = 0 (or minimal from budget floor)
	signals = health.ComputeSignals(0, 0, 0, "researching", 0)
	overall = health.ComputeOverallScore(signals)
	// Budget clarity gives 30 * 0.20 = 6, Decision Progress gives 20 * 0.30 = 6
	// So overall = 0*0.25 + 0*0.25 + 30*0.20 + 20*0.30 = 0 + 0 + 6 + 6 = 12
	if overall < 0 || overall > 20 {
		t.Errorf("expected overall in reasonable low range for empty mission, got %d", overall)
	}
}

// TestGrade verifies grade assignment thresholds.
func TestGrade(t *testing.T) {
	cases := []struct {
		score    int
		expected string
	}{
		{100, "A"},
		{90, "A"},
		{89, "B"},
		{80, "B"},
		{79, "C"},
		{70, "C"},
		{69, "D"},
		{60, "D"},
		{59, "F"},
		{0, "F"},
	}
	for _, tc := range cases {
		t.Run(tc.expected, func(t *testing.T) {
			got := health.ComputeGrade(tc.score)
			if got != tc.expected {
				t.Errorf("score=%d: expected grade %s, got %s", tc.score, tc.expected, got)
			}
		})
	}
}

// TestSignalFields verifies each signal has required fields.
func TestSignalFields(t *testing.T) {
	signals := health.ComputeSignals(3, 1, 50.0, "comparing", 2)
	for _, s := range signals {
		if s.Name == "" {
			t.Error("signal has empty Name")
		}
		if s.Description == "" {
			t.Errorf("signal %s has empty Description", s.Name)
		}
		if s.Icon == "" {
			t.Errorf("signal %s has empty Icon", s.Name)
		}
		if s.Score < 0 || s.Score > 100 {
			t.Errorf("signal %s score %d out of range [0,100]", s.Name, s.Score)
		}
		if s.Weight <= 0 {
			t.Errorf("signal %s has non-positive Weight %f", s.Name, s.Weight)
		}
	}
}

// TestComputeOverallScore_Weighted verifies partial weighted average.
func TestComputeOverallScore_Weighted(t *testing.T) {
	// 5 options (research=100, coverage=100), budget=0 (clarity=30), researching (progress=20)
	// expected = 100*0.25 + 100*0.25 + 30*0.20 + 20*0.30 = 25 + 25 + 6 + 6 = 62
	signals := health.ComputeSignals(5, 0, 0, "researching", 0)
	overall := health.ComputeOverallScore(signals)
	if overall != 62 {
		t.Errorf("expected overall 62, got %d", overall)
	}
}

// TestComputeSignals_BoundaryOneOption verifies 1 option gives 20 score for research/coverage.
func TestComputeSignals_BoundaryOneOption(t *testing.T) {
	signals := health.ComputeSignals(1, 0, 0, "researching", 0)
	research := findSignal(t, signals, "Research Depth")
	if research.Score != 20 {
		t.Errorf("expected Research Depth score 20 for 1 option, got %d", research.Score)
	}
	coverage := findSignal(t, signals, "Option Coverage")
	if coverage.Score != 20 {
		t.Errorf("expected Option Coverage score 20 for 1 option, got %d", coverage.Score)
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func findSignal(t *testing.T, signals []health.SignalScore, name string) health.SignalScore {
	t.Helper()
	for _, s := range signals {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("signal %q not found in signals %v", name, signals)
	return health.SignalScore{}
}
