package priceapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
)

// --- Mock repositories ---

type mockOptionRepo struct {
	opt *option.Option
	err error
}

func (m *mockOptionRepo) GetByID(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return m.opt, m.err
}
func (m *mockOptionRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Create(_ context.Context, o option.Option) (*option.Option, error) {
	return &o, nil
}
func (m *mockOptionRepo) Update(_ context.Context, _ uuid.UUID, _ option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Delete(_ context.Context, _ uuid.UUID) error           { return nil }
func (m *mockOptionRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error  { return nil }
func (m *mockOptionRepo) Pin(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Reject(_ context.Context, _ uuid.UUID, _ option.RejectRequest) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Unreject(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }

type mockMissionRepo struct {
	m   *mission.Mission
	err error
}

func (m *mockMissionRepo) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return m.m, m.err
}
func (m *mockMissionRepo) List(_ context.Context) ([]mission.Mission, error)  { return nil, nil }
func (m *mockMissionRepo) ListPaged(_ context.Context, _ *time.Time, _ int, _ bool) ([]mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) ListActive(_ context.Context) ([]mission.Mission, error) { return nil, nil }
func (m *mockMissionRepo) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) GetByShareToken(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) Create(_ context.Context, ms mission.Mission) (*mission.Mission, error) {
	return &ms, nil
}
func (m *mockMissionRepo) Update(_ context.Context, _ uuid.UUID, _ mission.UpdateRequest) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockMissionRepo) SetPhase(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockMissionRepo) SetShareToken(_ context.Context, _ uuid.UUID, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) ClearShareToken(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockMissionRepo) Archive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) Unarchive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) CloneMission(_ context.Context, _ string) (mission.Mission, error) {
	return mission.Mission{}, nil
}

// --- Helpers ---

// staticAggregator always returns the given result.
type staticAggregator struct {
	result *AggregatedResult
}

func (s *staticAggregator) Search(_ context.Context, _, _, _ string) (*AggregatedResult, error) {
	if s.result == nil {
		return &AggregatedResult{Listings: []PriceListing{}, Providers: []string{}}, nil
	}
	return s.result, nil
}

// buildRouter wires up the chi router the same way main.go would.
func buildRouter(h *Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/missions/{missionID}/options/{optionID}/live-prices", h.LivePrices)
	return r
}

func makeIDs() (missionID, optionID uuid.UUID) {
	return uuid.New(), uuid.New()
}

// --- Tests ---

func TestHandler_LivePrices_HappyPath(t *testing.T) {
	mID, oID := makeIDs()
	ms := &mission.Mission{ID: mID, Name: "Food Mission", Category: "food", Locale: "en"}
	opt := &option.Option{ID: oID, MissionID: mID, Name: "Cheese"}

	h := &Handler{
		aggregator:  &aggregatorAdapter{&staticAggregator{result: &AggregatedResult{
			Listings:  []PriceListing{{Source: "OFF", ProductName: "Cheese", Price: 3.50}},
			Providers: []string{"OpenFoodFacts"},
		}}},
		optionRepo:  &mockOptionRepo{opt: opt},
		missionRepo: &mockMissionRepo{m: ms},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+mID.String()+"/options/"+oID.String()+"/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	var result AggregatedResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result.Listings) != 1 {
		t.Errorf("expected 1 listing, got %d", len(result.Listings))
	}
}

func TestHandler_LivePrices_InvalidMissionID(t *testing.T) {
	h := &Handler{
		aggregator:  &aggregatorAdapter{&staticAggregator{}},
		optionRepo:  &mockOptionRepo{},
		missionRepo: &mockMissionRepo{},
	}
	oID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/api/missions/not-a-uuid/options/"+oID.String()+"/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_LivePrices_InvalidOptionID(t *testing.T) {
	mID := uuid.New()
	h := &Handler{
		aggregator:  &aggregatorAdapter{&staticAggregator{}},
		optionRepo:  &mockOptionRepo{},
		missionRepo: &mockMissionRepo{},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+mID.String()+"/options/not-a-uuid/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_LivePrices_OptionNotFound(t *testing.T) {
	mID, oID := makeIDs()
	h := &Handler{
		aggregator:  &aggregatorAdapter{&staticAggregator{}},
		optionRepo:  &mockOptionRepo{opt: nil},
		missionRepo: &mockMissionRepo{},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+mID.String()+"/options/"+oID.String()+"/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_LivePrices_MissionNotFound(t *testing.T) {
	mID, oID := makeIDs()
	opt := &option.Option{ID: oID, MissionID: mID, Name: "Widget"}
	h := &Handler{
		aggregator:  &aggregatorAdapter{&staticAggregator{}},
		optionRepo:  &mockOptionRepo{opt: opt},
		missionRepo: &mockMissionRepo{m: nil},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+mID.String()+"/options/"+oID.String()+"/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_LivePrices_OptionMissionMismatch(t *testing.T) {
	mID, oID := makeIDs()
	otherMissionID := uuid.New()
	// Option belongs to a different mission.
	opt := &option.Option{ID: oID, MissionID: otherMissionID, Name: "Widget"}
	ms := &mission.Mission{ID: mID, Category: "food"}
	h := &Handler{
		aggregator:  &aggregatorAdapter{&staticAggregator{}},
		optionRepo:  &mockOptionRepo{opt: opt},
		missionRepo: &mockMissionRepo{m: ms},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+mID.String()+"/options/"+oID.String()+"/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_LivePrices_EmptyListings(t *testing.T) {
	mID, oID := makeIDs()
	ms := &mission.Mission{ID: mID, Category: "electronics"}
	opt := &option.Option{ID: oID, MissionID: mID, Name: "Laptop"}
	h := &Handler{
		aggregator: &aggregatorAdapter{&staticAggregator{result: &AggregatedResult{
			Listings:  []PriceListing{},
			Providers: []string{},
		}}},
		optionRepo:  &mockOptionRepo{opt: opt},
		missionRepo: &mockMissionRepo{m: ms},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+mID.String()+"/options/"+oID.String()+"/live-prices", nil)
	w := httptest.NewRecorder()
	buildRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result AggregatedResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if result.Listings == nil {
		t.Error("expected non-nil empty listings array")
	}
	if len(result.Listings) != 0 {
		t.Errorf("expected 0 listings, got %d", len(result.Listings))
	}
}
