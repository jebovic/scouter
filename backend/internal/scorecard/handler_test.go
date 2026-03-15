package scorecard

import (
	"testing"
	"time"
)

func ptr[T any](v T) *T { return &v }

func TestBuildScorecardGradeA(t *testing.T) {
	now := time.Now()
	purchased := now.Add(3 * 24 * time.Hour)
	m := missionData{
		id:          "m1",
		name:        "MacBook Pro",
		createdAt:   now,
		completedAt: &purchased,
	}
	stats := missionStats{
		optionCount:   6,
		priceCheckCnt: 5,
		budget:        ptr(2000.0),
		purchasePrice: ptr(1800.0),
	}
	sc := buildScorecard(m, stats)
	if sc.Grade != "A" {
		t.Errorf("expected grade A, got %s (score=%d, dims=%+v)", sc.Grade, sc.Score, sc.Dimensions)
	}
	if sc.Score < 85 {
		t.Errorf("expected score >= 85, got %d", sc.Score)
	}
}

func TestBuildScorecardGradeD(t *testing.T) {
	old := time.Now().Add(-200 * 24 * time.Hour)
	purchased := time.Now()
	m := missionData{
		id:          "m2",
		name:        "Budget PC",
		createdAt:   old,
		completedAt: &purchased,
	}
	stats := missionStats{
		optionCount:   0,
		priceCheckCnt: 0,
		budget:        ptr(500.0),
		purchasePrice: ptr(600.0), // over budget
	}
	sc := buildScorecard(m, stats)
	if sc.Grade != "D" {
		t.Errorf("expected grade D, got %s (score=%d)", sc.Grade, sc.Score)
	}
}

func TestBuildScorecardNoCompletion(t *testing.T) {
	m := missionData{
		id:          "m3",
		name:        "Camera",
		createdAt:   time.Now(),
		completedAt: nil,
	}
	stats := missionStats{optionCount: 1, priceCheckCnt: 1}
	sc := buildScorecard(m, stats)
	if sc == nil {
		t.Fatal("expected non-nil scorecard")
	}
	if sc.MissionID != "m3" {
		t.Errorf("wrong missionId: %s", sc.MissionID)
	}
}

func TestGradeFor(t *testing.T) {
	tests := []struct {
		score    int
		expected string
	}{
		{90, "A"},
		{85, "A"},
		{80, "B"},
		{70, "B"},
		{60, "C"},
		{50, "C"},
		{40, "D"},
	}
	for _, tt := range tests {
		if got := gradeFor(tt.score); got != tt.expected {
			t.Errorf("gradeFor(%d) = %s, want %s", tt.score, got, tt.expected)
		}
	}
}

func TestAchievementsEarned(t *testing.T) {
	stats := missionStats{
		optionCount:   5,
		priceCheckCnt: 5,
	}
	dims := Dimensions{Research: 100, Pricing: 100, Budgeting: 100, Speed: 100}
	achievements := buildAchievements(stats, dims)

	earned := 0
	for _, a := range achievements {
		if a.Earned {
			earned++
		}
	}
	if earned != 4 {
		t.Errorf("expected all 4 achievements earned, got %d", earned)
	}
}

func TestAchievementsNoneEarned(t *testing.T) {
	stats := missionStats{}
	dims := Dimensions{}
	achievements := buildAchievements(stats, dims)
	for _, a := range achievements {
		if a.Earned {
			t.Errorf("expected no achievements earned, but %q is earned", a.Title)
		}
	}
}
