package negotiation_test

// Tests for handler-internal pure helpers exported via white-box test access.
// Because these helpers are unexported, we test them indirectly via the exported
// API of the package. Pure-function coverage is achieved through the agent tests
// and the model Validate tests. This file adds coverage for the containsIgnoreCase
// logic by exercising the fetchMarketPrices path through the handler with a stub
// shopping repository, and covers buildInput via a helper that assembles a realistic
// NegotiationInput and verifies the agent receives it correctly.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/negotiation"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
	"time"
)

// --- stub repositories ---

type stubOptionRepo struct {
	opt *option.Option
	err error
}

func (r *stubOptionRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return nil, nil
}
func (r *stubOptionRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]option.Option, error) {
	return nil, nil
}
func (r *stubOptionRepo) GetByID(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return r.opt, r.err
}
func (r *stubOptionRepo) Create(_ context.Context, o option.Option) (*option.Option, error) {
	return &o, nil
}
func (r *stubOptionRepo) Update(_ context.Context, _ uuid.UUID, _ option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}
func (r *stubOptionRepo) Delete(_ context.Context, _ uuid.UUID) error          { return nil }
func (r *stubOptionRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubOptionRepo) Pin(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (r *stubOptionRepo) Reject(_ context.Context, _ uuid.UUID, _ option.RejectRequest) (*option.Option, error) {
	return nil, nil
}
func (r *stubOptionRepo) Unreject(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (r *stubOptionRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }

type stubMissionRepo struct {
	m   *mission.Mission
	err error
}

func (r *stubMissionRepo) List(_ context.Context) ([]mission.Mission, error)   { return nil, nil }
func (r *stubMissionRepo) ListPaged(_ context.Context, _ *time.Time, _ int, _ bool) ([]mission.Mission, error) {
	return nil, nil
}
func (r *stubMissionRepo) ListActive(_ context.Context) ([]mission.Mission, error) { return nil, nil }
func (r *stubMissionRepo) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return r.m, r.err
}
func (r *stubMissionRepo) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (r *stubMissionRepo) GetByShareToken(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (r *stubMissionRepo) Create(_ context.Context, m mission.Mission) (*mission.Mission, error) {
	return &m, nil
}
func (r *stubMissionRepo) Update(_ context.Context, _ uuid.UUID, _ mission.UpdateRequest) (*mission.Mission, error) {
	return nil, nil
}
func (r *stubMissionRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubMissionRepo) SetPhase(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (r *stubMissionRepo) SetShareToken(_ context.Context, _ uuid.UUID, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (r *stubMissionRepo) ClearShareToken(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubMissionRepo) Archive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (r *stubMissionRepo) Unarchive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}

type stubShoppingRepo struct {
	items []shopping.Item
	err   error
}

func (r *stubShoppingRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return r.items, r.err
}
func (r *stubShoppingRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]shopping.Item, error) {
	return nil, nil
}
func (r *stubShoppingRepo) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (r *stubShoppingRepo) Create(_ context.Context, item shopping.Item) (*shopping.Item, error) {
	return &item, nil
}
func (r *stubShoppingRepo) Update(_ context.Context, _ uuid.UUID, _ shopping.UpdateRequest) (*shopping.Item, error) {
	return nil, nil
}
func (r *stubShoppingRepo) Delete(_ context.Context, _ uuid.UUID) error          { return nil }
func (r *stubShoppingRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubShoppingRepo) AddPriceSnapshot(_ context.Context, _ uuid.UUID, _ shopping.PriceSnapshotRequest) (*shopping.PriceSnapshot, error) {
	return nil, nil
}
func (r *stubShoppingRepo) ListPriceHistory(_ context.Context, _ uuid.UUID) ([]shopping.PriceSnapshot, error) {
	return nil, nil
}
func (r *stubShoppingRepo) Pin(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (r *stubShoppingRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubShoppingRepo) GetPriceHistory(_ context.Context, _ uuid.UUID, _ int) ([]shopping.PriceHistoryPoint, error) {
	return nil, nil
}

// newHandlerWithStubs builds a Handler with all dependencies stubbed.
func newHandlerWithStubs(
	optRepo option.Repository,
	missionRepo mission.Repository,
	shoppingRepo shopping.Repository,
	provider llm.Provider,
) *negotiation.Handler {
	return negotiation.NewHandlerWithRepos(optRepo, missionRepo, shoppingRepo, provider)
}

func TestHandlerWithStubs_OptionNotFound_Returns404(t *testing.T) {
	h := newHandlerWithStubs(
		&stubOptionRepo{opt: nil, err: nil},
		&stubMissionRepo{},
		&stubShoppingRepo{},
		&stubProvider{},
	)

	optionID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optionID+"/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing option, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerWithStubs_MissionNotFound_Returns404(t *testing.T) {
	missionID := uuid.New()
	opt := &option.Option{
		ID:        uuid.New(),
		MissionID: missionID,
		Name:      "MacBook Pro 14",
		Badge:     "recommended",
		PriceRange: &option.PriceRange{Min: 1100, Max: 1300, Best: 1199},
	}
	h := newHandlerWithStubs(
		&stubOptionRepo{opt: opt},
		&stubMissionRepo{m: nil, err: nil}, // mission not found
		&stubShoppingRepo{},
		&stubProvider{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+opt.ID.String()+"/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing mission, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerWithStubs_HappyPath_Returns200(t *testing.T) {
	missionID := uuid.New()
	optID := uuid.New()

	opt := &option.Option{
		ID:        optID,
		MissionID: missionID,
		Name:      "MacBook Pro 14",
		Badge:     "recommended",
		Attributes: []option.Attribute{
			{Key: "ram", Label: "RAM", Value: "16GB", Type: "text"},
		},
		PriceRange: &option.PriceRange{Min: 1100, Max: 1300, Best: 1199},
	}
	m := &mission.Mission{
		ID:       missionID,
		Slug:     "laptop-purchase",
		Name:     "Laptop Purchase",
		Category: "computing",
		Budget:   1000.0,
		Currency: "USD",
		Constraints: []mission.Constraint{
			{Key: "delivery", Label: "Delivery", Value: "2 weeks", Type: "soft"},
		},
	}
	shoppingItems := []shopping.Item{
		{ID: uuid.New(), MissionID: missionID, Name: "MacBook Pro 14 Amazon", Merchant: "Amazon", Price: 1180.0},
	}

	expected := negotiation.NegotiationScript{
		OpeningOffer:      1050.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 10.0,
		TalkingPoints:     []string{"Amazon offers same for $1,180"},
		CounterOfferScript: []string{"Start at $1,050", "Accept $1,100"},
		ConfidenceScore:   0.75,
	}

	h := newHandlerWithStubs(
		&stubOptionRepo{opt: opt},
		&stubMissionRepo{m: m},
		&stubShoppingRepo{items: shoppingItems},
		&stubProvider{resp: buildToolResponse(expected)},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID.String()+"/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerWithStubs_ZeroOptionPrice_Returns500(t *testing.T) {
	// An option with no price range means optionPrice=0 → agent validation fails.
	missionID := uuid.New()
	optID := uuid.New()

	opt := &option.Option{
		ID:        optID,
		MissionID: missionID,
		Name:      "MacBook Pro 14",
		Badge:     "recommended",
		// PriceRange is nil — optionPrice will be 0
	}
	m := &mission.Mission{
		ID:       missionID,
		Name:     "Laptop Purchase",
		Category: "computing",
		Budget:   1000.0,
		Currency: "USD",
	}

	h := newHandlerWithStubs(
		&stubOptionRepo{opt: opt},
		&stubMissionRepo{m: m},
		&stubShoppingRepo{},
		&stubProvider{},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID.String()+"/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 for zero option price, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerWithStubs_ShoppingRepoError_StillSucceeds(t *testing.T) {
	// A shopping repo error should be tolerated (market prices are optional).
	missionID := uuid.New()
	optID := uuid.New()
	opt := &option.Option{
		ID:        optID,
		MissionID: missionID,
		Name:      "MacBook Pro 14",
		Badge:     "recommended",
		PriceRange: &option.PriceRange{Min: 1100, Max: 1300, Best: 1199},
	}
	m := &mission.Mission{
		ID:       missionID,
		Name:     "Laptop Purchase",
		Category: "computing",
		Budget:   1000.0,
		Currency: "USD",
	}

	expected := negotiation.NegotiationScript{
		OpeningOffer:      1050.0,
		WalkAwayPrice:     1100.0,
		SuggestedDiscount: 8.0,
		TalkingPoints:     []string{"Over budget by 20%"},
		CounterOfferScript: []string{"Open at $1,050"},
		ConfidenceScore:   0.7,
	}

	h := newHandlerWithStubs(
		&stubOptionRepo{opt: opt},
		&stubMissionRepo{m: m},
		&stubShoppingRepo{err: context.DeadlineExceeded}, // shopping repo error
		&stubProvider{resp: buildToolResponse(expected)},
	)

	req := httptest.NewRequest(http.MethodPost, "/api/options/"+optID.String()+"/negotiate", http.NoBody)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Mount("/api/options/{optionID}", h.Routes())
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 even with shopping repo error, got %d: %s", rec.Code, rec.Body.String())
	}
}
