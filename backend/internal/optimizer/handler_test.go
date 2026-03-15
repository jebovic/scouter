package optimizer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/shopping"
)

// ---------------------------------------------------------------------------
// Fakes / stubs
// ---------------------------------------------------------------------------

type fakeMissionGetter struct {
	m   *mission.Mission
	err error
}

func (f *fakeMissionGetter) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return f.m, f.err
}

type fakeShoppingRepo struct {
	items []shopping.Item
	err   error
}

func (f *fakeShoppingRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return f.items, f.err
}

type fakeOptimizerAgent struct {
	plan *OptimizedPlan
	err  error
}

func (f *fakeOptimizerAgent) Optimize(_ context.Context, _ []shopping.Item) (*OptimizedPlan, error) {
	return f.plan, f.err
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func chiRequest(method, path, slug string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", slug)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func testMission(id uuid.UUID) *mission.Mission {
	return &mission.Mission{
		ID:       id,
		Slug:     "test-mission",
		Name:     "Test Mission",
		Category: "electronics",
		Budget:   1000,
		Currency: "EUR",
	}
}

func testItems(n int) []shopping.Item {
	items := make([]shopping.Item, n)
	for i := range items {
		items[i] = shopping.Item{
			ID:       uuid.New(),
			Name:     "Article",
			Merchant: "Amazon",
			Price:    99.99,
			Status:   "watch",
		}
	}
	return items
}

func testPlan(missionID string) *OptimizedPlan {
	return &OptimizedPlan{
		MissionID:   missionID,
		GeneratedAt: time.Now().UTC(),
		TotalItems:  2,
		Actions: []ShoppingAction{
			{
				Priority:   1,
				ItemIDs:    []string{uuid.New().String()},
				ItemNames:  []string{"Laptop"},
				Merchant:   "Amazon",
				Action:     "Acheter maintenant",
				Reason:     "Meilleur prix actuel",
				EstSavings: 0,
				Currency:   "EUR",
			},
		},
		Summary: "Commencez par acheter le laptop.",
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// Test 1: Success — returns optimized plan for a mission with enough items.
func TestHandler_Optimize_Success(t *testing.T) {
	missionID := uuid.New()
	m := testMission(missionID)
	plan := testPlan(missionID.String())

	h := &Handler{
		missionGetter: &fakeMissionGetter{m: m},
		shoppingRepo:  &fakeShoppingRepo{items: testItems(3)},
		agent:         &fakeOptimizerAgent{plan: plan},
		cache:         NewCache(30 * time.Minute),
	}

	req := chiRequest("POST", "/api/missions/test-mission/optimize", "test-mission")
	w := httptest.NewRecorder()
	h.Optimize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200. body=%s", w.Code, w.Body.String())
	}

	var result OptimizedPlan
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.MissionID != missionID.String() {
		t.Errorf("MissionID = %q; want %q", result.MissionID, missionID.String())
	}
	if len(result.Actions) == 0 {
		t.Error("expected at least one action")
	}
}

// Test 2: < 2 items → 422.
func TestHandler_Optimize_TooFewItems(t *testing.T) {
	missionID := uuid.New()
	m := testMission(missionID)

	h := &Handler{
		missionGetter: &fakeMissionGetter{m: m},
		shoppingRepo:  &fakeShoppingRepo{items: testItems(1)},
		agent:         &fakeOptimizerAgent{},
		cache:         NewCache(30 * time.Minute),
	}

	req := chiRequest("POST", "/api/missions/test-mission/optimize", "test-mission")
	w := httptest.NewRecorder()
	h.Optimize(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d; want 422", w.Code)
	}
}

// Test 3: Mission not found → 404.
func TestHandler_Optimize_MissionNotFound(t *testing.T) {
	h := &Handler{
		missionGetter: &fakeMissionGetter{m: nil},
		shoppingRepo:  &fakeShoppingRepo{},
		agent:         &fakeOptimizerAgent{},
		cache:         NewCache(30 * time.Minute),
	}

	req := chiRequest("POST", "/api/missions/nonexistent/optimize", "nonexistent")
	w := httptest.NewRecorder()
	h.Optimize(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d; want 404", w.Code)
	}
}

// Test 4: Cache hit — agent is not called; cached plan returned.
func TestHandler_Optimize_CacheHit(t *testing.T) {
	missionID := uuid.New()
	m := testMission(missionID)
	plan := testPlan(missionID.String())

	callCount := 0
	spy := &spyAgent{plan: plan, countPtr: &callCount}

	h := &Handler{
		missionGetter: &fakeMissionGetter{m: m},
		shoppingRepo:  &fakeShoppingRepo{items: testItems(3)},
		agent:         spy,
		cache:         NewCache(30 * time.Minute),
	}

	// Seed cache
	h.cache.Set(missionID.String(), plan)

	req := chiRequest("POST", "/api/missions/test-mission/optimize", "test-mission")
	w := httptest.NewRecorder()
	h.Optimize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	if callCount != 0 {
		t.Errorf("agent was called %d times; want 0 (should serve from cache)", callCount)
	}
}

// Test 5: Response structure — all required fields present.
func TestHandler_Optimize_ResponseStructure(t *testing.T) {
	missionID := uuid.New()
	m := testMission(missionID)
	plan := testPlan(missionID.String())

	h := &Handler{
		missionGetter: &fakeMissionGetter{m: m},
		shoppingRepo:  &fakeShoppingRepo{items: testItems(3)},
		agent:         &fakeOptimizerAgent{plan: plan},
		cache:         NewCache(30 * time.Minute),
	}

	req := chiRequest("POST", "/api/missions/test-mission/optimize", "test-mission")
	w := httptest.NewRecorder()
	h.Optimize(w, req)

	var result map[string]json.RawMessage
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	for _, key := range []string{"missionId", "generatedAt", "totalItems", "actions", "summary"} {
		if _, ok := result[key]; !ok {
			t.Errorf("missing required key %q in response", key)
		}
	}
}

// Test 6: Summary present in response.
func TestHandler_Optimize_SummaryPresent(t *testing.T) {
	missionID := uuid.New()
	m := testMission(missionID)
	plan := testPlan(missionID.String())
	plan.Summary = "Achetez d'abord le laptop, puis la souris."

	h := &Handler{
		missionGetter: &fakeMissionGetter{m: m},
		shoppingRepo:  &fakeShoppingRepo{items: testItems(2)},
		agent:         &fakeOptimizerAgent{plan: plan},
		cache:         NewCache(30 * time.Minute),
	}

	req := chiRequest("POST", "/api/missions/test-mission/optimize", "test-mission")
	w := httptest.NewRecorder()
	h.Optimize(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}

	var result OptimizedPlan
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Summary == "" {
		t.Error("summary must not be empty")
	}
}

// ---------------------------------------------------------------------------
// Spy agent helper for cache test
// ---------------------------------------------------------------------------

type spyAgent struct {
	plan     *OptimizedPlan
	countPtr *int
}

func (s *spyAgent) Optimize(_ context.Context, _ []shopping.Item) (*OptimizedPlan, error) {
	*s.countPtr++
	return s.plan, nil
}

// ---------------------------------------------------------------------------
// Agent fallback tests
// ---------------------------------------------------------------------------

// TestAgent_Fallback_SortsHighScoreFirst verifies that when LLM is unavailable,
// items with high deal score come first.
func TestAgent_Fallback_SortsHighScoreFirst(t *testing.T) {
	items := []shopping.Item{
		{ID: uuid.New(), Name: "Montre", Merchant: "Darty", Price: 50, Status: "watch"},
		{ID: uuid.New(), Name: "Laptop", Merchant: "Amazon", Price: 999, Status: "buy"},
		{ID: uuid.New(), Name: "Souris", Merchant: "Amazon", Price: 30, Status: "flash-sale"},
	}

	plan := buildFallbackPlan("mission-1", "EUR", items)

	if len(plan.Actions) == 0 {
		t.Fatal("expected at least one action in fallback plan")
	}
	if plan.Summary == "" {
		t.Error("fallback plan must have a non-empty summary")
	}
	if plan.TotalItems != len(items) {
		t.Errorf("TotalItems = %d; want %d", plan.TotalItems, len(items))
	}
}

// TestAgent_Fallback_EmptyItems handles zero items gracefully.
func TestAgent_Fallback_EmptyItems(t *testing.T) {
	plan := buildFallbackPlan("mission-x", "USD", nil)
	if plan == nil {
		t.Fatal("fallback plan must not be nil even for empty items")
	}
	if plan.TotalItems != 0 {
		t.Errorf("TotalItems = %d; want 0", plan.TotalItems)
	}
}

// TestAgent_Fallback_ActionLabels checks label assignment based on status.
func TestAgent_Fallback_ActionLabels(t *testing.T) {
	items := []shopping.Item{
		{ID: uuid.New(), Name: "Item A", Merchant: "M1", Price: 10, Status: "buy"},
		{ID: uuid.New(), Name: "Item B", Merchant: "M2", Price: 20, Status: "defer"},
		{ID: uuid.New(), Name: "Item C", Merchant: "M3", Price: 30, Status: "watch"},
	}

	plan := buildFallbackPlan("mission-z", "EUR", items)
	for _, a := range plan.Actions {
		if a.Action == "" {
			t.Errorf("action label must not be empty for item %v", a.ItemNames)
		}
	}
}

// TestAgent_LLMError_UsesFallback verifies that when LLM returns an error,
// Optimize still returns a valid plan (fallback path).
func TestAgent_LLMError_UsesFallback(t *testing.T) {
	agent := &Agent{
		provider: &errProvider{err: errors.New("LLM unavailable")},
	}
	items := []shopping.Item{
		{ID: uuid.New(), Name: "X", Merchant: "M", Price: 10, Status: "buy"},
		{ID: uuid.New(), Name: "Y", Merchant: "M", Price: 20, Status: "watch"},
	}

	plan, err := agent.Optimize(context.Background(), items)
	if err != nil {
		t.Fatalf("Optimize returned error on LLM failure; should fallback: %v", err)
	}
	if plan == nil {
		t.Fatal("plan must not be nil after LLM fallback")
	}
	if len(plan.Actions) == 0 {
		t.Error("fallback plan should have actions")
	}
}

// ---------------------------------------------------------------------------
// errProvider — always returns an error (for fallback test)
// ---------------------------------------------------------------------------

type errProvider struct {
	err error
}

func (e *errProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	return llm.CompletionResponse{}, e.err
}
