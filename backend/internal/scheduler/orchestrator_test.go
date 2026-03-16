package scheduler

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/shopping"
)

// ---- mock mission repository ----

type mockMissionRepo struct {
	missions []mission.Mission
	err      error
}

func (m *mockMissionRepo) ListActive(_ context.Context) ([]mission.Mission, error) {
	return m.missions, m.err
}

func (m *mockMissionRepo) List(_ context.Context) ([]mission.Mission, error) { return nil, nil }
func (m *mockMissionRepo) ListPaged(_ context.Context, _ *time.Time, _ int, _ bool) ([]mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) GetByShareToken(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) Create(_ context.Context, _ mission.Mission) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) Update(_ context.Context, _ uuid.UUID, _ mission.UpdateRequest) (*mission.Mission, error) {
	return nil, nil
}
func (m *mockMissionRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
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

func (m *mockMissionRepo) SetPhase(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (m *mockMissionRepo) CloneMission(_ context.Context, _ string) (mission.Mission, error) {
	return mission.Mission{}, nil
}

// ---- mock option repository ----

type mockOptionRepo struct {
	opts []option.Option
	err  error
}

func (m *mockOptionRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return m.opts, m.err
}
func (m *mockOptionRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) GetByID(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Create(_ context.Context, _ option.Option) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Update(_ context.Context, _ uuid.UUID, _ option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}
func (m *mockOptionRepo) Delete(_ context.Context, _ uuid.UUID) error          { return nil }
func (m *mockOptionRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
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
func (m *mockOptionRepo) SaveTranslations(_ context.Context, _ uuid.UUID, _ string, _ json.RawMessage) error {
	return nil
}

// trackingOptionRepo records which mission IDs were queried.
type trackingOptionRepo struct {
	mockOptionRepo
	onList func(uuid.UUID)
}

func (t *trackingOptionRepo) ListByMission(_ context.Context, id uuid.UUID) ([]option.Option, error) {
	t.onList(id)
	return t.opts, t.err
}

// ---- mock shopping repository ----

type mockShoppingRepo struct {
	snapshots []shopping.PriceSnapshot
	err       error
}

func (m *mockShoppingRepo) ListPriceHistory(_ context.Context, _ uuid.UUID) ([]shopping.PriceSnapshot, error) {
	return m.snapshots, m.err
}
func (m *mockShoppingRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return nil, nil
}
func (m *mockShoppingRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]shopping.Item, error) {
	return nil, nil
}
func (m *mockShoppingRepo) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (m *mockShoppingRepo) Create(_ context.Context, _ shopping.Item) (*shopping.Item, error) {
	return nil, nil
}
func (m *mockShoppingRepo) Update(_ context.Context, _ uuid.UUID, _ shopping.UpdateRequest) (*shopping.Item, error) {
	return nil, nil
}
func (m *mockShoppingRepo) Delete(_ context.Context, _ uuid.UUID) error          { return nil }
func (m *mockShoppingRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockShoppingRepo) AddPriceSnapshot(_ context.Context, _ uuid.UUID, _ shopping.PriceSnapshotRequest) (*shopping.PriceSnapshot, error) {
	return nil, nil
}
func (m *mockShoppingRepo) Pin(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (m *mockShoppingRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockShoppingRepo) GetPriceHistory(_ context.Context, _ uuid.UUID, _ int) ([]shopping.PriceHistoryPoint, error) {
	return nil, nil
}

// ---- mock notification repository ----

type mockOrchestratorNotifRepo struct {
	created         []notification.CreateRequest
	existing        []notification.Notification // returned by ListFiltered for dedup tests
	createErr       error                        // injected error for Create calls
	listFilteredErr error                        // injected error for ListFiltered calls
}

func (m *mockOrchestratorNotifRepo) Create(_ context.Context, req notification.CreateRequest) (*notification.Notification, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.created = append(m.created, req)
	return &notification.Notification{}, nil
}
func (m *mockOrchestratorNotifRepo) List(_ context.Context, _ int) ([]notification.Notification, error) {
	return nil, nil
}
func (m *mockOrchestratorNotifRepo) UnreadCount(_ context.Context) (int, error)    { return 0, nil }
func (m *mockOrchestratorNotifRepo) MarkRead(_ context.Context, _ uuid.UUID) error { return nil }
func (m *mockOrchestratorNotifRepo) MarkAllRead(_ context.Context) error           { return nil }
func (m *mockOrchestratorNotifRepo) Delete(_ context.Context, _ uuid.UUID) error   { return nil }
func (m *mockOrchestratorNotifRepo) ListFiltered(_ context.Context, _ notification.ListFilter) ([]notification.Notification, error) {
	if m.listFilteredErr != nil {
		return nil, m.listFilteredErr
	}
	return m.existing, nil
}

// ---- mock pricing runners ----

type mockPricingRunner struct {
	items []shopping.Item
	err   error
}

func (m *mockPricingRunner) Run(_ context.Context, _ mission.Mission, _ []option.Option, _ *pricing.FeedbackInput) ([]shopping.Item, error) {
	return m.items, m.err
}

// countingPricingRunner tracks Run invocations.
type countingPricingRunner struct {
	calls int
}

func (c *countingPricingRunner) Run(_ context.Context, _ mission.Mission, _ []option.Option, _ *pricing.FeedbackInput) ([]shopping.Item, error) {
	c.calls++
	return nil, nil
}

// ---- helpers ----

func newTestOrchestrator(
	missionRepo mission.Repository,
	optionRepo option.Repository,
	shoppingRepo shopping.Repository,
	notifRepo notification.Repository,
	agent pricingRunner,
	maxMissions int,
) *Orchestrator {
	return NewOrchestrator(
		missionRepo,
		optionRepo,
		shoppingRepo,
		notifRepo,
		agent,
		maxMissions,
		slog.New(slog.NewTextHandler(os.Stderr, nil)),
	)
}

func orchTestMission() mission.Mission {
	return mission.Mission{ID: uuid.New(), Slug: "test-mission"}
}

// snapshotsOldestFirst builds price history snapshots with oldest entry first.
// Prices are assigned in the order supplied; the first price is oldest.
// To produce a TrendDropping score, supply prices in decreasing order (e.g. 100, 95, 88).
func snapshotsOldestFirst(prices ...float64) []shopping.PriceSnapshot {
	base := time.Now().Add(-time.Duration(len(prices)) * time.Hour)
	snaps := make([]shopping.PriceSnapshot, len(prices))
	for i, p := range prices {
		snaps[i] = shopping.PriceSnapshot{
			ID:         uuid.New(),
			Price:      p,
			RecordedAt: base.Add(time.Duration(i) * time.Hour),
		}
	}
	return snaps
}

// ---- tests ----

func TestOrchestrator_RunPriceChecks_NoMissions(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	o := newTestOrchestrator(
		&mockMissionRepo{missions: nil},
		&mockOptionRepo{},
		&mockShoppingRepo{},
		notifRepo,
		&mockPricingRunner{},
		10,
	)
	o.RunPriceChecks(context.Background())
	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications, got %d", len(notifRepo.created))
	}
}

func TestOrchestrator_RunPriceChecks_MaxMissionsCap(t *testing.T) {
	missions := make([]mission.Mission, 5)
	for i := range missions {
		missions[i] = mission.Mission{ID: uuid.New(), Slug: "mission"}
	}

	var processed []uuid.UUID
	optRepo := &trackingOptionRepo{onList: func(id uuid.UUID) { processed = append(processed, id) }}

	o := newTestOrchestrator(
		&mockMissionRepo{missions: missions},
		optRepo,
		&mockShoppingRepo{},
		&mockOrchestratorNotifRepo{},
		&mockPricingRunner{items: nil},
		3, // cap at 3 even though 5 missions exist
	)
	o.RunPriceChecks(context.Background())

	if len(processed) != 3 {
		t.Errorf("processed %d missions; want 3", len(processed))
	}
}

func TestOrchestrator_MaybeNotify_TargetHit(t *testing.T) {
	target := 100.0
	item := shopping.Item{
		ID:          uuid.New(),
		Name:        "Widget",
		Price:       80.0, // 80 < 100 target → PctBelowTarget = 20%
		TargetPrice: &target,
	}
	notifRepo := &mockOrchestratorNotifRepo{}
	o := newTestOrchestrator(
		&mockMissionRepo{},
		&mockOptionRepo{},
		// oldest→newest: 95, 90, 85 — dropping trend, avg=90 > 80 current
		&mockShoppingRepo{snapshots: snapshotsOldestFirst(95, 90, 85)},
		notifRepo,
		&mockPricingRunner{},
		10,
	)

	if err := o.maybeNotify(context.Background(), orchTestMission(), item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifRepo.created))
	}
	if notifRepo.created[0].Type != notification.TypeTargetHit {
		t.Errorf("type = %q; want %q", notifRepo.created[0].Type, notification.TypeTargetHit)
	}
}

func TestOrchestrator_MaybeNotify_PriceDrop(t *testing.T) {
	item := shopping.Item{
		ID:    uuid.New(),
		Name:  "Widget",
		Price: 80.0, // no target — only price drop path
	}
	notifRepo := &mockOrchestratorNotifRepo{}
	o := newTestOrchestrator(
		&mockMissionRepo{},
		&mockOptionRepo{},
		// oldest→newest: 100, 95, 88 — dropping (-12%), avg≈94.3 > 80 current
		&mockShoppingRepo{snapshots: snapshotsOldestFirst(100, 95, 88)},
		notifRepo,
		&mockPricingRunner{},
		10,
	)

	if err := o.maybeNotify(context.Background(), orchTestMission(), item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifRepo.created))
	}
	if notifRepo.created[0].Type != notification.TypePriceDrop {
		t.Errorf("type = %q; want %q", notifRepo.created[0].Type, notification.TypePriceDrop)
	}
}

func TestOrchestrator_MaybeNotify_InsufficientHistory(t *testing.T) {
	target := 100.0
	item := shopping.Item{ID: uuid.New(), Name: "Widget", Price: 80.0, TargetPrice: &target}
	notifRepo := &mockOrchestratorNotifRepo{}
	o := newTestOrchestrator(
		&mockMissionRepo{},
		&mockOptionRepo{},
		// only 2 snapshots — below minSnapshotsForScore(3), ComputeDealScore returns nil
		&mockShoppingRepo{snapshots: snapshotsOldestFirst(95, 90)},
		notifRepo,
		&mockPricingRunner{},
		10,
	)

	if err := o.maybeNotify(context.Background(), orchTestMission(), item); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications with insufficient history, got %d", len(notifRepo.created))
	}
}

func TestOrchestrator_RunPriceChecks_ContextCancelled(t *testing.T) {
	missions := []mission.Mission{
		{ID: uuid.New(), Slug: "m1"},
		{ID: uuid.New(), Slug: "m2"},
	}
	agent := &countingPricingRunner{}

	// Provide real options so checkMission would reach pricingAgent.Run if the
	// context check were absent — ensuring we test the ctx guard, not the empty-opts guard.
	optRepo := &mockOptionRepo{opts: []option.Option{{ID: uuid.New()}}}

	o := newTestOrchestrator(
		&mockMissionRepo{missions: missions},
		optRepo,
		&mockShoppingRepo{},
		&mockOrchestratorNotifRepo{},
		agent,
		10,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before RunPriceChecks so ctx.Err() != nil on first loop iteration

	o.RunPriceChecks(ctx)

	if agent.calls != 0 {
		t.Errorf("expected 0 pricing agent calls with cancelled context, got %d", agent.calls)
	}
}
