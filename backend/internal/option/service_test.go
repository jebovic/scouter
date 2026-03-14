package option

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type mockRepo struct {
	options map[uuid.UUID]*Option
	nextErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{options: make(map[uuid.UUID]*Option)}
}

func (m *mockRepo) ListByMissionPaged(_ context.Context, missionID uuid.UUID, _ *time.Time, limit int) ([]Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	var out []Option
	for _, o := range m.options {
		if o.MissionID == missionID {
			out = append(out, *o)
		}
	}
	if len(out) > limit+1 {
		out = out[:limit+1]
	}
	return out, nil
}

func (m *mockRepo) ListByMission(_ context.Context, missionID uuid.UUID) ([]Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	var out []Option
	for _, o := range m.options {
		if o.MissionID == missionID {
			out = append(out, *o)
		}
	}
	return out, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	o, ok := m.options[id]
	if !ok {
		return nil, nil
	}
	return o, nil
}

func (m *mockRepo) Create(_ context.Context, o Option) (*Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	o.ID = uuid.New()
	m.options[o.ID] = &o
	return &o, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, req UpdateRequest) (*Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	o, ok := m.options[id]
	if !ok {
		return nil, nil
	}
	if req.Badge != nil {
		o.Badge = *req.Badge
	}
	return o, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	delete(m.options, id)
	return nil
}

func (m *mockRepo) DeleteByMission(_ context.Context, missionID uuid.UUID) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	for id, o := range m.options {
		if o.MissionID == missionID {
			delete(m.options, id)
		}
	}
	return nil
}

func (m *mockRepo) Pin(_ context.Context, id uuid.UUID) (*Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	o, ok := m.options[id]
	if !ok {
		return nil, nil
	}
	o.Pinned = true
	return o, nil
}

func (m *mockRepo) Reject(_ context.Context, id uuid.UUID, req RejectRequest) (*Option, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	o, ok := m.options[id]
	if !ok {
		return nil, nil
	}
	o.Rejected = true
	o.RejectReason = req.Reason
	return o, nil
}

func (m *mockRepo) DeletePinned(_ context.Context, missionID uuid.UUID) error {
	if m.nextErr != nil {
		return m.nextErr
	}
	for id, o := range m.options {
		if o.MissionID == missionID && o.Pinned {
			delete(m.options, id)
		}
	}
	return nil
}

func TestService_Create_RequiresName(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: ""})
	if err == nil {
		t.Error("Create with empty name should return error")
	}
}

func TestService_Create_DefaultBadge(t *testing.T) {
	svc := NewService(newMockRepo())
	o, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: "MacBook Pro"})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if o.Badge != "watch" {
		t.Errorf("Badge = %q; want %q", o.Badge, "watch")
	}
}

func TestService_Create_PreservesBadge(t *testing.T) {
	svc := NewService(newMockRepo())
	o, err := svc.Create(context.Background(), uuid.New(), CreateRequest{
		Name:  "Dell XPS",
		Badge: "recommended",
	})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if o.Badge != "recommended" {
		t.Errorf("Badge = %q; want %q", o.Badge, "recommended")
	}
}

func TestService_Create_NilSlicesBecomEmpty(t *testing.T) {
	svc := NewService(newMockRepo())
	o, err := svc.Create(context.Background(), uuid.New(), CreateRequest{Name: "ThinkPad"})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if o.Attributes == nil {
		t.Error("Attributes = nil; want empty slice")
	}
	if o.Warnings == nil {
		t.Error("Warnings = nil; want empty slice")
	}
}

func TestService_ListByMission_NilToEmpty(t *testing.T) {
	svc := NewService(newMockRepo())
	opts, err := svc.ListByMission(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListByMission() error: %v", err)
	}
	if opts == nil {
		t.Error("ListByMission() = nil; want empty slice")
	}
}
