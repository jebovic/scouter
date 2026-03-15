package persona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// ---------------------------------------------------------------------------
// Stub repos
// ---------------------------------------------------------------------------

type stubShoppingMissionRepo struct {
	missions []mission.Mission
	err      error
}

func (s *stubShoppingMissionRepo) List(_ context.Context) ([]mission.Mission, error) {
	return s.missions, s.err
}
func (s *stubShoppingMissionRepo) ListPaged(_ context.Context, _ *time.Time, _ int, _ bool) ([]mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) ListActive(_ context.Context) ([]mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) GetByShareToken(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) Create(_ context.Context, m mission.Mission) (*mission.Mission, error) {
	return &m, nil
}
func (s *stubShoppingMissionRepo) Update(_ context.Context, _ uuid.UUID, _ mission.UpdateRequest) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubShoppingMissionRepo) SetPhase(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (s *stubShoppingMissionRepo) SetShareToken(_ context.Context, _ uuid.UUID, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) ClearShareToken(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubShoppingMissionRepo) Archive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) Unarchive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubShoppingMissionRepo) CloneMission(_ context.Context, _ string) (mission.Mission, error) {
	return mission.Mission{}, nil
}

type stubShoppingOptionRepo struct {
	opts []option.Option
	err  error
}

func (s *stubShoppingOptionRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return s.opts, s.err
}
func (s *stubShoppingOptionRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]option.Option, error) {
	return nil, nil
}
func (s *stubShoppingOptionRepo) GetByID(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (s *stubShoppingOptionRepo) Create(_ context.Context, o option.Option) (*option.Option, error) {
	return &o, nil
}
func (s *stubShoppingOptionRepo) Update(_ context.Context, _ uuid.UUID, _ option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}
func (s *stubShoppingOptionRepo) Delete(_ context.Context, _ uuid.UUID) error       { return nil }
func (s *stubShoppingOptionRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubShoppingOptionRepo) Pin(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (s *stubShoppingOptionRepo) Reject(_ context.Context, _ uuid.UUID, _ option.RejectRequest) (*option.Option, error) {
	return nil, nil
}
func (s *stubShoppingOptionRepo) Unreject(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (s *stubShoppingOptionRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }

type stubShoppingItemRepo struct {
	items []shopping.Item
	err   error
}

func (s *stubShoppingItemRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return s.items, s.err
}
func (s *stubShoppingItemRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) Create(_ context.Context, item shopping.Item) (*shopping.Item, error) {
	return &item, nil
}
func (s *stubShoppingItemRepo) Update(_ context.Context, _ uuid.UUID, _ shopping.UpdateRequest) (*shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) Delete(_ context.Context, _ uuid.UUID) error       { return nil }
func (s *stubShoppingItemRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubShoppingItemRepo) AddPriceSnapshot(_ context.Context, _ uuid.UUID, _ shopping.PriceSnapshotRequest) (*shopping.PriceSnapshot, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) ListPriceHistory(_ context.Context, _ uuid.UUID) ([]shopping.PriceSnapshot, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) GetPriceHistory(_ context.Context, _ uuid.UUID, _ int) ([]shopping.PriceHistoryPoint, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) Pin(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingItemRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }

// stubShoppingLLM returns a canned analyze_shopping_persona tool call response.
type stubShoppingLLM struct {
	resp llm.CompletionResponse
	err  error
}

func (s *stubShoppingLLM) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	return s.resp, s.err
}

func validShoppingLLMResponse() llm.CompletionResponse {
	return llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{
				Name: "analyze_shopping_persona",
				Input: map[string]any{
					"archetype":   "Le Chasseur de Deals",
					"icon":        "🎯",
					"description": "Vous êtes obsédé par les meilleures offres.",
					"strengths":   []any{"Économise beaucoup", "Très patient", "Analyse les prix"},
					"blindspots":  []any{"Peut manquer de qualité", "Perd du temps"},
					"tips":        []any{"Utilisez des alertes de prix", "Comparez toujours", "Attendez les soldes"},
					"scoreProfile": map[string]any{
						"frugalité":    float64(90),
						"impulsivité":  float64(15),
						"recherche":    float64(85),
						"budget":       float64(80),
						"qualité":      float64(60),
					},
				},
			},
		},
	}
}

func makeMissions(count int) []mission.Mission {
	missions := make([]mission.Mission, count)
	for i := range missions {
		missions[i] = mission.Mission{
			ID:       uuid.New(),
			Category: "electronics",
			Budget:   500,
		}
	}
	return missions
}

// ---------------------------------------------------------------------------
// Test 1: stub returned when fewer than 2 missions
// ---------------------------------------------------------------------------

func TestGetShoppingPersona_StubWhenFewMissions(t *testing.T) {
	mRepo := &stubShoppingMissionRepo{missions: makeMissions(1)}
	oRepo := &stubShoppingOptionRepo{}
	sRepo := &stubShoppingItemRepo{}
	agent := NewShoppingPersonaAgent(nil)
	h := NewShoppingPersonaHandler(mRepo, oRepo, sRepo, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/persona/shopping", nil)
	w := httptest.NewRecorder()
	h.GetShoppingPersona(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var p ShoppingPersona
	if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.Archetype == "" {
		t.Error("expected non-empty archetype in stub persona")
	}
}

// ---------------------------------------------------------------------------
// Test 2: success with enough missions — archetype from known list
// ---------------------------------------------------------------------------

func TestGetShoppingPersona_SuccessWithData(t *testing.T) {
	mRepo := &stubShoppingMissionRepo{missions: makeMissions(3)}
	oRepo := &stubShoppingOptionRepo{}
	sRepo := &stubShoppingItemRepo{
		items: []shopping.Item{{Price: 100}, {Price: 250}},
	}
	agent := NewShoppingPersonaAgent(&stubShoppingLLM{resp: validShoppingLLMResponse()})
	h := NewShoppingPersonaHandler(mRepo, oRepo, sRepo, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/persona/shopping", nil)
	w := httptest.NewRecorder()
	h.GetShoppingPersona(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var p ShoppingPersona
	if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if p.Archetype != "Le Chasseur de Deals" {
		t.Errorf("expected archetype 'Le Chasseur de Deals', got %q", p.Archetype)
	}
}

// ---------------------------------------------------------------------------
// Test 3: archetype value is one of the known French archetypes
// ---------------------------------------------------------------------------

func TestGetShoppingPersona_ArchetypeFromKnownList(t *testing.T) {
	mRepo := &stubShoppingMissionRepo{missions: makeMissions(3)}
	oRepo := &stubShoppingOptionRepo{}
	sRepo := &stubShoppingItemRepo{}
	agent := NewShoppingPersonaAgent(&stubShoppingLLM{resp: validShoppingLLMResponse()})
	h := NewShoppingPersonaHandler(mRepo, oRepo, sRepo, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/persona/shopping", nil)
	w := httptest.NewRecorder()
	h.GetShoppingPersona(w, req)

	var p ShoppingPersona
	if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}

	known := map[string]bool{
		"Le Chasseur de Deals": true,
		"Le Chercheur":         true,
		"L'Impulsif":           true,
		"Le Budgetaire":        true,
		"Le Premium":           true,
		"Le Prudent":           true,
		// stub default archetype also valid
		"Nouveau Scouter": true,
	}
	if !known[p.Archetype] {
		t.Errorf("archetype %q is not in the known list", p.Archetype)
	}
}

// ---------------------------------------------------------------------------
// Test 4: score profile keys present
// ---------------------------------------------------------------------------

func TestGetShoppingPersona_ScoreProfileKeysPresent(t *testing.T) {
	mRepo := &stubShoppingMissionRepo{missions: makeMissions(3)}
	oRepo := &stubShoppingOptionRepo{}
	sRepo := &stubShoppingItemRepo{}
	agent := NewShoppingPersonaAgent(&stubShoppingLLM{resp: validShoppingLLMResponse()})
	h := NewShoppingPersonaHandler(mRepo, oRepo, sRepo, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/persona/shopping", nil)
	w := httptest.NewRecorder()
	h.GetShoppingPersona(w, req)

	var p ShoppingPersona
	if err := json.NewDecoder(w.Body).Decode(&p); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(p.ScoreProfile) == 0 {
		t.Error("expected non-empty score profile")
	}
}

// ---------------------------------------------------------------------------
// Test 5: cache hit — second call returns same persona without re-calling LLM
// ---------------------------------------------------------------------------

func TestGetShoppingPersona_CacheHit(t *testing.T) {
	mRepo := &stubShoppingMissionRepo{missions: makeMissions(3)}
	oRepo := &stubShoppingOptionRepo{}
	sRepo := &stubShoppingItemRepo{}

	callCount := 0
	llmProvider := &countingShoppingLLM{inner: &stubShoppingLLM{resp: validShoppingLLMResponse()}, count: &callCount}
	agent := NewShoppingPersonaAgent(llmProvider)
	h := NewShoppingPersonaHandler(mRepo, oRepo, sRepo, agent)

	// First call — populates cache.
	req1 := httptest.NewRequest(http.MethodGet, "/api/persona/shopping", nil)
	w1 := httptest.NewRecorder()
	h.GetShoppingPersona(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first call: expected 200, got %d", w1.Code)
	}

	// Second call — should hit cache, not re-invoke LLM.
	req2 := httptest.NewRequest(http.MethodGet, "/api/persona/shopping", nil)
	w2 := httptest.NewRecorder()
	h.GetShoppingPersona(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("second call: expected 200, got %d", w2.Code)
	}
	if callCount != 1 {
		t.Errorf("expected LLM called once (cache hit on second), got %d calls", callCount)
	}
}

type countingShoppingLLM struct {
	inner llm.Provider
	count *int
}

func (c *countingShoppingLLM) Complete(ctx context.Context, req llm.CompletionRequest) (llm.CompletionResponse, error) {
	*c.count++
	return c.inner.Complete(ctx, req)
}
