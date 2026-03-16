package coach

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// MockMissionService returns mission by slug.
type MockMissionService struct {
	missions map[string]*mission.Mission
}

func (m *MockMissionService) GetBySlug(ctx context.Context, slug string) (*mission.Mission, error) {
	if ms, ok := m.missions[slug]; ok {
		return ms, nil
	}
	return nil, nil
}

// MockCoachAgent returns predefined tips.
type MockCoachAgent struct {
	tips []CoachTip
	err  error
}

func (m *MockCoachAgent) Run(ctx context.Context, missionID mission.Mission, optionsCount, itemsCount int) ([]CoachTip, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tips, nil
}

// MockOptionRepository provides options for a mission.
type MockOptionRepository struct {
	optionsByMission map[uuid.UUID][]option.Option
}

func (m *MockOptionRepository) ListByMission(ctx context.Context, missionID uuid.UUID) ([]option.Option, error) {
	return m.optionsByMission[missionID], nil
}

func (m *MockOptionRepository) ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]option.Option, error) {
	return m.optionsByMission[missionID], nil
}

func (m *MockOptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*option.Option, error) {
	return nil, nil
}

func (m *MockOptionRepository) Create(ctx context.Context, o option.Option) (*option.Option, error) {
	return nil, nil
}

func (m *MockOptionRepository) Update(ctx context.Context, id uuid.UUID, req option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}

func (m *MockOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockOptionRepository) DeleteByMission(ctx context.Context, missionID uuid.UUID) error {
	return nil
}

func (m *MockOptionRepository) Pin(ctx context.Context, id uuid.UUID) (*option.Option, error) {
	return nil, nil
}

func (m *MockOptionRepository) Reject(ctx context.Context, id uuid.UUID, req option.RejectRequest) (*option.Option, error) {
	return nil, nil
}

func (m *MockOptionRepository) Unreject(ctx context.Context, id uuid.UUID) (*option.Option, error) {
	return nil, nil
}

func (m *MockOptionRepository) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	return nil
}

// MockShoppingRepository provides items for a mission.
type MockShoppingRepository struct {
	itemsByMission map[uuid.UUID][]shopping.Item
}

func (m *MockShoppingRepository) ListByMission(ctx context.Context, missionID uuid.UUID) ([]shopping.Item, error) {
	return m.itemsByMission[missionID], nil
}

func (m *MockShoppingRepository) ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]shopping.Item, error) {
	return m.itemsByMission[missionID], nil
}

func (m *MockShoppingRepository) GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}

func (m *MockShoppingRepository) Create(ctx context.Context, item shopping.Item) (*shopping.Item, error) {
	return nil, nil
}

func (m *MockShoppingRepository) Update(ctx context.Context, id uuid.UUID, req shopping.UpdateRequest) (*shopping.Item, error) {
	return nil, nil
}

func (m *MockShoppingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockShoppingRepository) DeleteByMission(ctx context.Context, missionID uuid.UUID) error {
	return nil
}

func (m *MockShoppingRepository) AddPriceSnapshot(ctx context.Context, itemID uuid.UUID, req shopping.PriceSnapshotRequest) (*shopping.PriceSnapshot, error) {
	return nil, nil
}

func (m *MockShoppingRepository) ListPriceHistory(ctx context.Context, itemID uuid.UUID) ([]shopping.PriceSnapshot, error) {
	return nil, nil
}

func (m *MockShoppingRepository) GetPriceHistory(ctx context.Context, itemID uuid.UUID, limit int) ([]shopping.PriceHistoryPoint, error) {
	return nil, nil
}

func (m *MockShoppingRepository) Pin(ctx context.Context, id uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}

func (m *MockShoppingRepository) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	return nil
}

func TestHandler_GetCoachTips_Success(t *testing.T) {
	missionID := uuid.New()
	m := &mission.Mission{
		ID:       missionID,
		Slug:     "test-mission",
		Name:     "Test Mission",
		Category: "electronics",
		Budget:   1000,
		Currency: "USD",
		Phase:    "researching",
	}

	mockTips := []CoachTip{
		{
			Category: "research",
			Priority: "high",
			Title:    "Expand search",
			Body:     "Consider more brands",
		},
	}

	handler := &Handler{
		agent: &MockCoachAgent{tips: mockTips},
		missionGetter: &MockMissionService{
			missions: map[string]*mission.Mission{
				"test-mission": m,
			},
		},
		optionRepo: &MockOptionRepository{
			optionsByMission: map[uuid.UUID][]option.Option{missionID: make([]option.Option, 3)},
		},
		shoppingRepo: &MockShoppingRepository{
			itemsByMission: map[uuid.UUID][]shopping.Item{missionID: make([]shopping.Item, 5)},
		},
		cache: NewCache(15 * time.Minute),
	}

	req := httptest.NewRequest("GET", "/api/missions/test-mission/coach", nil)
	// Set the slug URL parameter manually
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"slug"}, Values: []string{"test-mission"}},
	}))
	w := httptest.NewRecorder()

	handler.getCoachTips(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", w.Code)
	}

	var result CoachResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if result.MissionID != missionID.String() {
		t.Errorf("MissionID = %q; want %q", result.MissionID, missionID.String())
	}
	if len(result.Tips) != 1 {
		t.Errorf("len(Tips) = %d; want 1", len(result.Tips))
	}
}

func TestHandler_GetCoachTips_MissionNotFound(t *testing.T) {
	handler := &Handler{
		agent:         &MockCoachAgent{},
		missionGetter: &MockMissionService{missions: map[string]*mission.Mission{}},
		optionRepo:    &MockOptionRepository{optionsByMission: map[uuid.UUID][]option.Option{}},
		shoppingRepo:  &MockShoppingRepository{itemsByMission: map[uuid.UUID][]shopping.Item{}},
		cache:         NewCache(15 * time.Minute),
	}

	req := httptest.NewRequest("GET", "/api/missions/nonexistent/coach", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"slug"}, Values: []string{"nonexistent"}},
	}))
	w := httptest.NewRecorder()

	handler.getCoachTips(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d; want 404", w.Code)
	}
}

func TestHandler_GetCoachTips_AgentError_ReturnsDegradedDTO(t *testing.T) {
	missionID := uuid.New()
	m := &mission.Mission{
		ID:       missionID,
		Slug:     "broken-mission",
		Name:     "Broken Mission",
		Category: "electronics",
		Budget:   500,
		Currency: "EUR",
	}

	handler := &Handler{
		agent: &MockCoachAgent{err: fmt.Errorf("llm unavailable")},
		missionGetter: &MockMissionService{
			missions: map[string]*mission.Mission{
				"broken-mission": m,
			},
		},
		optionRepo:   &MockOptionRepository{optionsByMission: map[uuid.UUID][]option.Option{}},
		shoppingRepo: &MockShoppingRepository{itemsByMission: map[uuid.UUID][]shopping.Item{}},
		cache:        NewCache(15 * time.Minute),
	}

	req := httptest.NewRequest("GET", "/api/missions/broken-mission/coach", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"slug"}, Values: []string{"broken-mission"}},
	}))
	w := httptest.NewRecorder()

	handler.getCoachTips(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want 200 (degraded)", w.Code)
	}

	var result CoachResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode degraded response: %v", err)
	}
	if len(result.Tips) == 0 {
		t.Error("expected at least one placeholder tip in degraded response")
	}
	if result.MissionID != missionID.String() {
		t.Errorf("MissionID = %q; want %q", result.MissionID, missionID.String())
	}
}

func TestHandler_GetCoachTips_CacheHit(t *testing.T) {
	missionID := uuid.New()
	m := &mission.Mission{
		ID:       missionID,
		Slug:     "cached-mission",
		Name:     "Cached Mission",
		Category: "electronics",
		Budget:   1000,
		Currency: "USD",
	}

	cachedTips := []CoachTip{
		{
			ID:       "cached123",
			Category: "budget",
			Priority: "medium",
			Title:    "Cached tip",
			Body:     "This came from cache",
		},
	}

	handler := &Handler{
		agent: &MockCoachAgent{tips: cachedTips},
		missionGetter: &MockMissionService{
			missions: map[string]*mission.Mission{
				"cached-mission": m,
			},
		},
		optionRepo: &MockOptionRepository{
			optionsByMission: map[uuid.UUID][]option.Option{missionID: make([]option.Option, 3)},
		},
		shoppingRepo: &MockShoppingRepository{
			itemsByMission: map[uuid.UUID][]shopping.Item{missionID: make([]shopping.Item, 5)},
		},
		cache: NewCache(15 * time.Minute),
	}

	// Seed the cache
	handler.cache.Set(missionID, CoachResponse{
		MissionID:   missionID.String(),
		Tips:        cachedTips,
		GeneratedAt: time.Now().Add(-1 * time.Minute),
	})

	req := httptest.NewRequest("GET", "/api/missions/cached-mission/coach", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, &chi.Context{
		URLParams: chi.RouteParams{Keys: []string{"slug"}, Values: []string{"cached-mission"}},
	}))
	w := httptest.NewRecorder()

	handler.getCoachTips(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", w.Code)
	}

	var result CoachResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(result.Tips) == 0 {
		t.Fatal("no tips returned")
	}
	if result.Tips[0].ID != "cached123" {
		t.Errorf("Got tip from cache; tip ID mismatch")
	}
}
