package rebalancer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/mission"
)

// ---------------------------------------------------------------------------
// Fakes
// ---------------------------------------------------------------------------

type fakeEnvelopeRepo struct {
	envelopes []envelope.Envelope
	err       error
}

func (f *fakeEnvelopeRepo) FindAll(_ context.Context) ([]envelope.Envelope, error) {
	return f.envelopes, f.err
}

type fakeMissionRepo struct {
	missions []mission.Mission
	err      error
}

func (f *fakeMissionRepo) List(_ context.Context) ([]mission.Mission, error) {
	return f.missions, f.err
}

type fakeRebalancerAgent struct {
	plan *RebalancePlan
	err  error
}

func (f *fakeRebalancerAgent) Rebalance(_ context.Context, _ []envelope.Envelope, _ []mission.Mission) (*RebalancePlan, error) {
	return f.plan, f.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newHandler(envRepo envelopeRepo, msnRepo missionLister, agent agentRunner) *Handler {
	return &Handler{
		envelopeRepo: envRepo,
		missionRepo:  msnRepo,
		agent:        agent,
		cache:        NewCache(6 * time.Hour),
	}
}

func testEnvelopes() []envelope.Envelope {
	return []envelope.Envelope{
		{
			ID:            uuid.New(),
			Name:          "Alimentation",
			MonthlyAmount: 250,
			Currency:      "EUR",
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			Name:          "Transport",
			MonthlyAmount: 100,
			Currency:      "EUR",
			CreatedAt:     time.Now(),
		},
	}
}

func testMissions() []mission.Mission {
	envID := uuid.New()
	return []mission.Mission{
		{
			ID:       uuid.New(),
			Slug:     "courses-janvier",
			Name:     "Courses Janvier",
			Budget:   300,
			Currency: "EUR",
			Phase:    "done",
			EnvelopeID: &envID,
		},
	}
}

func testPlan(envelopes []envelope.Envelope) *RebalancePlan {
	adjustments := make([]EnvelopeAdjustment, 0, len(envelopes))
	var totalCurrent, totalSuggested float64
	for _, env := range envelopes {
		suggested := env.MonthlyAmount * 1.10
		delta := ((suggested - env.MonthlyAmount) / env.MonthlyAmount) * 100
		adjustments = append(adjustments, EnvelopeAdjustment{
			EnvelopeID:       env.ID.String(),
			EnvelopeName:     env.Name,
			CurrentMonthly:   env.MonthlyAmount,
			SuggestedMonthly: suggested,
			DeltaPercent:     delta,
			Reason:           "Vous dépensez plus que prévu dans cette catégorie.",
		})
		totalCurrent += env.MonthlyAmount
		totalSuggested += suggested
	}
	return &RebalancePlan{
		Adjustments:    adjustments,
		Summary:        "Votre budget sous-estime vos dépenses réelles.",
		TotalCurrent:   totalCurrent,
		TotalSuggested: totalSuggested,
		CachedAt:       time.Now().Unix(),
	}
}

func doGet(h *Handler) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/budget/rebalance", nil)
	w := httptest.NewRecorder()
	h.Get(w, req)
	return w
}

// ---------------------------------------------------------------------------
// Test 1: No envelopes → returns empty plan (200)
// ---------------------------------------------------------------------------

func TestHandler_Get_NoEnvelopes(t *testing.T) {
	h := newHandler(
		&fakeEnvelopeRepo{envelopes: []envelope.Envelope{}},
		&fakeMissionRepo{missions: []mission.Mission{}},
		&fakeRebalancerAgent{},
	)

	w := doGet(h)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200. body = %s", w.Code, w.Body.String())
	}

	var plan RebalancePlan
	if err := json.NewDecoder(w.Body).Decode(&plan); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(plan.Adjustments) != 0 {
		t.Errorf("expected 0 adjustments for empty envelopes, got %d", len(plan.Adjustments))
	}
}

// ---------------------------------------------------------------------------
// Test 2: Success with adjustments
// ---------------------------------------------------------------------------

func TestHandler_Get_SuccessWithAdjustments(t *testing.T) {
	envs := testEnvelopes()
	missions := testMissions()
	plan := testPlan(envs)

	h := newHandler(
		&fakeEnvelopeRepo{envelopes: envs},
		&fakeMissionRepo{missions: missions},
		&fakeRebalancerAgent{plan: plan},
	)

	w := doGet(h)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200. body = %s", w.Code, w.Body.String())
	}

	var got RebalancePlan
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Adjustments) != 2 {
		t.Errorf("adjustments = %d; want 2", len(got.Adjustments))
	}
	if got.Summary == "" {
		t.Error("summary must not be empty")
	}
	if got.TotalCurrent <= 0 {
		t.Error("totalCurrent must be positive")
	}
}

// ---------------------------------------------------------------------------
// Test 3: Cache hit — agent is not called a second time
// ---------------------------------------------------------------------------

func TestHandler_Get_CacheHit(t *testing.T) {
	envs := testEnvelopes()
	plan := testPlan(envs)

	callCount := 0
	spy := &spyAgent{plan: plan, countPtr: &callCount}

	h := &Handler{
		envelopeRepo: &fakeEnvelopeRepo{envelopes: envs},
		missionRepo:  &fakeMissionRepo{missions: testMissions()},
		agent:        spy,
		cache:        NewCache(6 * time.Hour),
	}

	// First request — populates cache
	w1 := doGet(h)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request status = %d", w1.Code)
	}
	if callCount != 1 {
		t.Fatalf("agent called %d times after first request; want 1", callCount)
	}

	// Second request — must hit cache, not call agent again
	w2 := doGet(h)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request status = %d", w2.Code)
	}
	if callCount != 1 {
		t.Errorf("agent called %d times; want 1 (second should use cache)", callCount)
	}
}

// ---------------------------------------------------------------------------
// Test 4: Agent error → 500
// ---------------------------------------------------------------------------

func TestHandler_Get_AgentError(t *testing.T) {
	h := newHandler(
		&fakeEnvelopeRepo{envelopes: testEnvelopes()},
		&fakeMissionRepo{missions: testMissions()},
		&fakeRebalancerAgent{err: errors.New("LLM unavailable")},
	)

	w := doGet(h)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want 500", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Test 5: Multi-currency — plan returned, totals computed
// ---------------------------------------------------------------------------

func TestHandler_Get_MultiCurrency(t *testing.T) {
	envs := []envelope.Envelope{
		{ID: uuid.New(), Name: "Alimentation", MonthlyAmount: 300, Currency: "EUR", CreatedAt: time.Now()},
		{ID: uuid.New(), Name: "Transport US",  MonthlyAmount: 200, Currency: "USD", CreatedAt: time.Now()},
	}
	plan := testPlan(envs)

	h := newHandler(
		&fakeEnvelopeRepo{envelopes: envs},
		&fakeMissionRepo{missions: []mission.Mission{}},
		&fakeRebalancerAgent{plan: plan},
	)

	w := doGet(h)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200. body = %s", w.Code, w.Body.String())
	}

	var got RebalancePlan
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The handler must still return a valid plan even with mixed currencies.
	if len(got.Adjustments) != 2 {
		t.Errorf("adjustments = %d; want 2", len(got.Adjustments))
	}
}

// ---------------------------------------------------------------------------
// Spy agent
// ---------------------------------------------------------------------------

type spyAgent struct {
	plan     *RebalancePlan
	countPtr *int
}

func (s *spyAgent) Rebalance(_ context.Context, _ []envelope.Envelope, _ []mission.Mission) (*RebalancePlan, error) {
	*s.countPtr++
	return s.plan, nil
}
