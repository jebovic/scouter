package comparison

import (
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// Weight model validation tests (no DB required)
// ---------------------------------------------------------------------------

func TestWeight_ValidFields(t *testing.T) {
	w := Weight{
		ID:        "uuid-1",
		MissionID: "mission-uuid-1",
		Criteria:  "price",
		Weight:    75,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if w.Criteria != "price" {
		t.Errorf("expected criteria 'price', got %q", w.Criteria)
	}
	if w.Weight != 75 {
		t.Errorf("expected weight 75, got %d", w.Weight)
	}
}

func TestWeight_ZeroIsValid(t *testing.T) {
	w := Weight{
		ID:        "uuid-2",
		MissionID: "mission-uuid-2",
		Criteria:  "warranty",
		Weight:    0,
	}
	if w.Weight != 0 {
		t.Errorf("weight 0 should be valid, got %d", w.Weight)
	}
}

func TestWeight_MaxIsValid(t *testing.T) {
	w := Weight{
		ID:        "uuid-3",
		MissionID: "mission-uuid-3",
		Criteria:  "quality",
		Weight:    100,
	}
	if w.Weight != 100 {
		t.Errorf("weight 100 should be valid, got %d", w.Weight)
	}
}

func TestSetWeightRequest_Fields(t *testing.T) {
	req := SetWeightRequest{
		Criteria: "brand",
		Weight:   50,
	}
	if req.Criteria != "brand" {
		t.Errorf("expected criteria 'brand', got %q", req.Criteria)
	}
	if req.Weight != 50 {
		t.Errorf("expected weight 50, got %d", req.Weight)
	}
}

// ---------------------------------------------------------------------------
// OptionScore model tests
// ---------------------------------------------------------------------------

func TestOptionScore_Fields(t *testing.T) {
	s := OptionScore{
		OptionID:   "opt-uuid-1",
		OptionName: "Sony WH-1000XM5",
		Score:      87.5,
		Rank:       1,
	}
	if s.Score != 87.5 {
		t.Errorf("expected score 87.5, got %f", s.Score)
	}
	if s.Rank != 1 {
		t.Errorf("expected rank 1, got %d", s.Rank)
	}
}

func TestOptionScore_ZeroScore(t *testing.T) {
	s := OptionScore{
		OptionID:   "opt-uuid-2",
		OptionName: "Unknown Option",
		Score:      0,
		Rank:       3,
	}
	if s.Score != 0 {
		t.Errorf("expected score 0, got %f", s.Score)
	}
}

// ---------------------------------------------------------------------------
// MatrixResponse model tests
// ---------------------------------------------------------------------------

func TestMatrixResponse_EmptyWeightsAndScores(t *testing.T) {
	resp := MatrixResponse{
		Weights: []Weight{},
		Scores:  []OptionScore{},
	}
	if len(resp.Weights) != 0 {
		t.Errorf("expected 0 weights, got %d", len(resp.Weights))
	}
	if len(resp.Scores) != 0 {
		t.Errorf("expected 0 scores, got %d", len(resp.Scores))
	}
}

func TestMatrixResponse_WithData(t *testing.T) {
	resp := MatrixResponse{
		Weights: []Weight{
			{ID: "w1", MissionID: "m1", Criteria: "price", Weight: 80},
			{ID: "w2", MissionID: "m1", Criteria: "quality", Weight: 60},
		},
		Scores: []OptionScore{
			{OptionID: "o1", OptionName: "Option A", Score: 90.0, Rank: 1},
			{OptionID: "o2", OptionName: "Option B", Score: 70.0, Rank: 2},
		},
	}
	if len(resp.Weights) != 2 {
		t.Errorf("expected 2 weights, got %d", len(resp.Weights))
	}
	if len(resp.Scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(resp.Scores))
	}
	if resp.Scores[0].Rank != 1 {
		t.Errorf("first score should have rank 1, got %d", resp.Scores[0].Rank)
	}
}

// ---------------------------------------------------------------------------
// Weight validation helper tests (business logic)
// ---------------------------------------------------------------------------

func TestValidateWeight_InRange(t *testing.T) {
	cases := []struct {
		weight  int
		wantErr bool
	}{
		{0, false},
		{50, false},
		{100, false},
		{-1, true},
		{101, true},
	}
	for _, tc := range cases {
		err := validateWeight(tc.weight)
		if tc.wantErr && err == nil {
			t.Errorf("weight %d: expected error, got nil", tc.weight)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("weight %d: unexpected error: %v", tc.weight, err)
		}
	}
}

func TestValidateCriteria_NotEmpty(t *testing.T) {
	cases := []struct {
		criteria string
		wantErr  bool
	}{
		{"price", false},
		{"quality", false},
		{"warranty", false},
		{"", true},
		{"   ", true},
	}
	for _, tc := range cases {
		err := validateCriteria(tc.criteria)
		if tc.wantErr && err == nil {
			t.Errorf("criteria %q: expected error, got nil", tc.criteria)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("criteria %q: unexpected error: %v", tc.criteria, err)
		}
	}
}

// ---------------------------------------------------------------------------
// Score computation tests (pure logic, no DB)
// ---------------------------------------------------------------------------

func TestComputeScores_NoWeights(t *testing.T) {
	opts := []rawOption{
		{id: "o1", name: "Option A", attrs: map[string]interface{}{"price": 500.0}},
	}
	weights := []Weight{}
	scores := computeScores(opts, weights)
	if len(scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(scores))
	}
	if scores[0].Score != 0 {
		t.Errorf("expected score 0 with no weights, got %f", scores[0].Score)
	}
}

func TestComputeScores_PriceInverted(t *testing.T) {
	// Lower price should yield a higher score (inverted).
	opts := []rawOption{
		{id: "o1", name: "Cheap", attrs: map[string]interface{}{"price": 100.0}},
		{id: "o2", name: "Expensive", attrs: map[string]interface{}{"price": 900.0}},
	}
	weights := []Weight{
		{Criteria: "price", Weight: 100},
	}
	scores := computeScores(opts, weights)
	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}
	// Find cheap and expensive scores
	var cheapScore, expensiveScore float64
	for _, s := range scores {
		if s.OptionID == "o1" {
			cheapScore = s.Score
		} else {
			expensiveScore = s.Score
		}
	}
	if cheapScore <= expensiveScore {
		t.Errorf("cheap option should score higher than expensive: cheap=%f, expensive=%f", cheapScore, expensiveScore)
	}
}

func TestComputeScores_QualityHigherIsBetter(t *testing.T) {
	// Higher quality score should yield higher matrix score.
	opts := []rawOption{
		{id: "o1", name: "Low Quality", attrs: map[string]interface{}{"quality": 20.0}},
		{id: "o2", name: "High Quality", attrs: map[string]interface{}{"quality": 90.0}},
	}
	weights := []Weight{
		{Criteria: "quality", Weight: 100},
	}
	scores := computeScores(opts, weights)
	var lowScore, highScore float64
	for _, s := range scores {
		if s.OptionID == "o1" {
			lowScore = s.Score
		} else {
			highScore = s.Score
		}
	}
	if highScore <= lowScore {
		t.Errorf("high-quality option should score higher: high=%f, low=%f", highScore, lowScore)
	}
}

func TestComputeScores_MissingAttribute(t *testing.T) {
	// If option lacks the criterion key, it contributes 0 for that criterion.
	opts := []rawOption{
		{id: "o1", name: "Option A", attrs: map[string]interface{}{}},
	}
	weights := []Weight{
		{Criteria: "quality", Weight: 100},
	}
	scores := computeScores(opts, weights)
	if len(scores) != 1 {
		t.Fatalf("expected 1 score, got %d", len(scores))
	}
	if scores[0].Score != 0 {
		t.Errorf("expected 0 score for missing attribute, got %f", scores[0].Score)
	}
}

func TestComputeScores_NonNumericAttribute(t *testing.T) {
	// Non-numeric attribute value should be skipped (contributes 0).
	opts := []rawOption{
		{id: "o1", name: "Option A", attrs: map[string]interface{}{"quality": "excellent"}},
	}
	weights := []Weight{
		{Criteria: "quality", Weight: 100},
	}
	scores := computeScores(opts, weights)
	if scores[0].Score != 0 {
		t.Errorf("non-numeric attribute should contribute 0, got %f", scores[0].Score)
	}
}

func TestComputeScores_MultiCriteria(t *testing.T) {
	// Verify weighted combination: one cheap/low-quality vs one expensive/high-quality.
	opts := []rawOption{
		{id: "o1", name: "Budget", attrs: map[string]interface{}{"price": 100.0, "quality": 30.0}},
		{id: "o2", name: "Premium", attrs: map[string]interface{}{"price": 900.0, "quality": 90.0}},
	}
	// Equal weights: price (inverted) and quality. Budget should win on price but lose on quality.
	weights := []Weight{
		{Criteria: "price", Weight: 50},
		{Criteria: "quality", Weight: 50},
	}
	scores := computeScores(opts, weights)
	if len(scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(scores))
	}
	// Both scores should be between 0 and 100
	for _, s := range scores {
		if s.Score < 0 || s.Score > 100 {
			t.Errorf("score %f out of range [0,100] for option %s", s.Score, s.OptionName)
		}
	}
}

func TestComputeScores_RankAssignment(t *testing.T) {
	opts := []rawOption{
		{id: "o1", name: "Low", attrs: map[string]interface{}{"quality": 10.0}},
		{id: "o2", name: "Mid", attrs: map[string]interface{}{"quality": 50.0}},
		{id: "o3", name: "High", attrs: map[string]interface{}{"quality": 90.0}},
	}
	weights := []Weight{
		{Criteria: "quality", Weight: 100},
	}
	scores := computeScores(opts, weights)
	// Find rank 1 — should be the high quality option
	var rank1ID string
	for _, s := range scores {
		if s.Rank == 1 {
			rank1ID = s.OptionID
		}
	}
	if rank1ID != "o3" {
		t.Errorf("expected rank 1 to be 'o3' (high quality), got %q", rank1ID)
	}
}

func TestComputeScores_AllSameScore(t *testing.T) {
	// When all options have the same value, ranks should still be assigned sequentially.
	opts := []rawOption{
		{id: "o1", name: "A", attrs: map[string]interface{}{"quality": 50.0}},
		{id: "o2", name: "B", attrs: map[string]interface{}{"quality": 50.0}},
	}
	weights := []Weight{
		{Criteria: "quality", Weight: 100},
	}
	scores := computeScores(opts, weights)
	ranks := map[int]bool{}
	for _, s := range scores {
		if ranks[s.Rank] {
			t.Errorf("duplicate rank %d", s.Rank)
		}
		ranks[s.Rank] = true
	}
}

func TestComputeScores_EmptyOptions(t *testing.T) {
	scores := computeScores([]rawOption{}, []Weight{{Criteria: "price", Weight: 50}})
	if len(scores) != 0 {
		t.Errorf("expected 0 scores for empty options, got %d", len(scores))
	}
}

func TestNormalizeValue_Price(t *testing.T) {
	// Price: lower is better → higher value gets inverted to lower normalized score
	cases := []struct {
		criteria string
		val      float64
		min      float64
		max      float64
		wantHigh bool // whether this value should normalize high (>50)
	}{
		{"price", 100, 100, 1000, true},   // lowest price → highest score
		{"price", 1000, 100, 1000, false}, // highest price → lowest score
		{"quality", 90, 0, 100, true},     // high quality → high score
		{"quality", 10, 0, 100, false},    // low quality → low score
	}
	for _, tc := range cases {
		got := normalizeValue(tc.criteria, tc.val, tc.min, tc.max)
		if tc.wantHigh && got < 50 {
			t.Errorf("criteria=%s val=%f: expected normalized>50, got %f", tc.criteria, tc.val, got)
		}
		if !tc.wantHigh && got > 50 {
			t.Errorf("criteria=%s val=%f: expected normalized<50, got %f", tc.criteria, tc.val, got)
		}
	}
}

func TestNormalizeValue_AllSame(t *testing.T) {
	// When min == max, normalization should return 50 (midpoint) to avoid division by zero.
	got := normalizeValue("quality", 50, 50, 50)
	if got != 50 {
		t.Errorf("when min==max, expected 50, got %f", got)
	}
}
