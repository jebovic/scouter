package shopping

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockRepo struct {
	items     map[uuid.UUID]*Item
	snapshots []PriceSnapshot
	nextErr   error
}

func newMockRepo() *mockRepo {
	return &mockRepo{items: make(map[uuid.UUID]*Item)}
}

func (m *mockRepo) ListByMissionPaged(_ context.Context, missionID uuid.UUID, _ *time.Time, limit int) ([]Item, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	var out []Item
	for _, it := range m.items {
		if it.MissionID == missionID {
			out = append(out, *it)
		}
	}
	if len(out) > limit+1 {
		out = out[:limit+1]
	}
	return out, nil
}

func (m *mockRepo) ListByMission(_ context.Context, missionID uuid.UUID) ([]Item, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	var out []Item
	for _, it := range m.items {
		if it.MissionID == missionID {
			out = append(out, *it)
		}
	}
	return out, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*Item, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	it, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	return it, nil
}

func (m *mockRepo) Create(_ context.Context, item Item) (*Item, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	item.ID = uuid.New()
	m.items[item.ID] = &item
	return &item, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, req UpdateRequest) (*Item, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	it, ok := m.items[id]
	if !ok {
		return nil, nil
	}
	if req.Status != nil {
		it.Status = *req.Status
	}
	return it, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	delete(m.items, id)
	return nil
}

func (m *mockRepo) DeleteByMission(_ context.Context, missionID uuid.UUID) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	for id, it := range m.items {
		if it.MissionID == missionID {
			delete(m.items, id)
		}
	}
	return nil
}

func (m *mockRepo) AddPriceSnapshot(_ context.Context, itemID uuid.UUID, req PriceSnapshotRequest) (*PriceSnapshot, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	snap := PriceSnapshot{ID: uuid.New(), ItemID: itemID, Price: req.Price, Note: req.Note}
	m.snapshots = append(m.snapshots, snap)
	return &snap, nil
}

func (m *mockRepo) ListPriceHistory(_ context.Context, itemID uuid.UUID) ([]PriceSnapshot, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	var out []PriceSnapshot
	for _, s := range m.snapshots {
		if s.ItemID == itemID {
			out = append(out, s)
		}
	}
	return out, nil
}

func TestService_Create_RequiresName(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: "", Price: 10})
	if err == nil {
		t.Error("Create with empty name should return error")
	}
}

func TestService_Create_RejectsNegativePrice(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: "GPU", Price: -1})
	if err == nil {
		t.Error("Create with negative price should return error")
	}
}

func TestService_Create_AllowsZeroPrice(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: "Free item", Price: 0})
	if err != nil {
		t.Errorf("Create with zero price should be allowed, got: %v", err)
	}
}

func TestService_Create_DefaultStatus(t *testing.T) {
	svc := NewService(newMockRepo())
	item, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: "GPU", Price: 500})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if item.Status != "watch" {
		t.Errorf("Status = %q; want %q", item.Status, "watch")
	}
}

func TestService_Create_PreservesStatus(t *testing.T) {
	svc := NewService(newMockRepo())
	item, err := svc.Create(context.Background(), uuid.New(), CreateRequest{
		Name:   "GPU",
		Price:  500,
		Status: "buy",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if item.Status != "buy" {
		t.Errorf("Status = %q; want %q", item.Status, "buy")
	}
}

func TestService_AddPriceSnapshot_RejectsNegativePrice(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.AddPriceSnapshot(context.Background(), uuid.New(), PriceSnapshotRequest{Price: -5})
	if err == nil {
		t.Error("AddPriceSnapshot with negative price should return error")
	}
}

func TestService_AddPriceSnapshot_AllowsZeroPrice(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.AddPriceSnapshot(context.Background(), uuid.New(), PriceSnapshotRequest{Price: 0})
	if err != nil {
		t.Errorf("AddPriceSnapshot with zero price should be allowed, got: %v", err)
	}
}

func TestService_ListByMission_NilToEmpty(t *testing.T) {
	svc := NewService(newMockRepo())
	items, err := svc.ListByMission(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListByMission() error: %v", err)
	}
	if items == nil {
		t.Error("ListByMission() = nil; want empty slice")
	}
}

func TestService_ListPriceHistory_NilToEmpty(t *testing.T) {
	svc := NewService(newMockRepo())
	snaps, err := svc.ListPriceHistory(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListPriceHistory() error: %v", err)
	}
	if snaps == nil {
		t.Error("ListPriceHistory() = nil; want empty slice")
	}
}
