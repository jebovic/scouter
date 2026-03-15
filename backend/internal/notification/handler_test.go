package notification

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
	"github.com/jackc/pgx/v5"
)

// mockNotifRepo implements Repository for testing.
// Per-method error fields allow each test to inject failures independently
// without cross-method bleed.
type mockNotifRepo struct {
	notifs         []Notification
	listErr        error
	unreadCountErr error
	markReadErr    error
	markAllReadErr error
}

func (m *mockNotifRepo) Create(_ context.Context, req CreateRequest) (*Notification, error) {
	n := Notification{
		ID:        uuid.New(),
		MissionID: req.MissionID,
		ItemID:    req.ItemID,
		Type:      req.Type,
		Title:     req.Title,
		Body:      req.Body,
		CreatedAt: time.Now(),
	}
	m.notifs = append(m.notifs, n)
	return &n, nil
}

func (m *mockNotifRepo) List(_ context.Context, limit int) ([]Notification, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	out := m.notifs
	if len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *mockNotifRepo) UnreadCount(_ context.Context) (int, error) {
	if m.unreadCountErr != nil {
		return 0, m.unreadCountErr
	}
	count := 0
	for _, n := range m.notifs {
		if !n.Read {
			count++
		}
	}
	return count, nil
}

func (m *mockNotifRepo) MarkRead(_ context.Context, id uuid.UUID) error {
	if m.markReadErr != nil {
		return m.markReadErr
	}
	for i, n := range m.notifs {
		if n.ID == id {
			m.notifs[i].Read = true
			return nil
		}
	}
	return pgx.ErrNoRows
}

func (m *mockNotifRepo) MarkAllRead(_ context.Context) error {
	if m.markAllReadErr != nil {
		return m.markAllReadErr
	}
	for i := range m.notifs {
		m.notifs[i].Read = true
	}
	return nil
}

func (m *mockNotifRepo) ListFiltered(_ context.Context, f ListFilter) ([]Notification, error) {
	return m.List(context.Background(), f.Limit)
}

func (m *mockNotifRepo) Delete(_ context.Context, id uuid.UUID) error {
	for i, n := range m.notifs {
		if n.ID == id {
			m.notifs = append(m.notifs[:i], m.notifs[i+1:]...)
			return nil
		}
	}
	return pgx.ErrNoRows
}

func newHandler(repo Repository) http.Handler {
	h := NewHandler(repo)
	r := chi.NewRouter()
	r.Mount("/", h.Routes())
	return r
}

// ---- list tests ----

func TestHandler_List_ReturnsNotifications(t *testing.T) {
	repo := &mockNotifRepo{
		notifs: []Notification{
			{ID: uuid.New(), Type: TypePriceDrop, Title: "Drop", CreatedAt: time.Now()},
			{ID: uuid.New(), Type: TypeTargetHit, Title: "Hit", CreatedAt: time.Now()},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	newHandler(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var got []Notification
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d; want 2", len(got))
	}
}

func TestHandler_List_RespectsLimitParam(t *testing.T) {
	repo := &mockNotifRepo{}
	for i := 0; i < 5; i++ {
		repo.notifs = append(repo.notifs, Notification{ID: uuid.New(), CreatedAt: time.Now()})
	}
	req := httptest.NewRequest(http.MethodGet, "/?limit=2", nil)
	w := httptest.NewRecorder()
	newHandler(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var got []Notification
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d; want 2", len(got))
	}
}

func TestHandler_List_LimitClamped(t *testing.T) {
	repo := &mockNotifRepo{}
	for i := 0; i < 250; i++ {
		repo.notifs = append(repo.notifs, Notification{ID: uuid.New(), CreatedAt: time.Now()})
	}
	req := httptest.NewRequest(http.MethodGet, "/?limit=9999", nil)
	w := httptest.NewRecorder()
	newHandler(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var got []Notification
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != maxNotifLimit {
		t.Errorf("len = %d; want %d (maxNotifLimit)", len(got), maxNotifLimit)
	}
}

func TestHandler_List_InvalidLimitFallsBackToDefault(t *testing.T) {
	cases := []string{"?limit=abc", "?limit=0", "?limit=-1"}
	for _, q := range cases {
		repo := &mockNotifRepo{}
		for i := 0; i < defaultNotifLimit+10; i++ {
			repo.notifs = append(repo.notifs, Notification{ID: uuid.New(), CreatedAt: time.Now()})
		}
		req := httptest.NewRequest(http.MethodGet, "/"+q, nil)
		w := httptest.NewRecorder()
		newHandler(repo).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("[%s] status = %d; want 200", q, w.Code)
		}
		var got []Notification
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("[%s] unmarshal: %v", q, err)
		}
		if len(got) != defaultNotifLimit {
			t.Errorf("[%s] len = %d; want %d (defaultNotifLimit)", q, len(got), defaultNotifLimit)
		}
	}
}

func TestHandler_List_EmptyReturnsArray(t *testing.T) {
	repo := &mockNotifRepo{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	newHandler(repo).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	if body := w.Body.String(); body != "[]\n" {
		t.Errorf("body = %q; want \"[]\\n\"", body)
	}
}

func TestHandler_List_RepoError(t *testing.T) {
	repo := &mockNotifRepo{listErr: errors.New("db error")}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	newHandler(repo).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want 500", w.Code)
	}
}

// ---- mark-read tests ----

func TestHandler_MarkRead(t *testing.T) {
	existingID := uuid.New()

	tests := []struct {
		name       string
		repoSetup  func() *mockNotifRepo
		id         uuid.UUID
		wantStatus int
	}{
		{
			name: "success",
			repoSetup: func() *mockNotifRepo {
				return &mockNotifRepo{
					notifs: []Notification{{ID: existingID, Type: TypePriceDrop, CreatedAt: time.Now()}},
				}
			},
			id:         existingID,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "not found",
			repoSetup:  func() *mockNotifRepo { return &mockNotifRepo{} }, // empty — no notifications
			id:         uuid.New(),
			wantStatus: http.StatusNotFound,
		},
		{
			name: "repo error",
			repoSetup: func() *mockNotifRepo {
				return &mockNotifRepo{markReadErr: errors.New("db error")}
			},
			id:         uuid.New(),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := tc.repoSetup()
			req := httptest.NewRequest(http.MethodPatch, "/"+tc.id.String()+"/read", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.id.String())
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()

			h := NewHandler(repo)
			h.markRead(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d; want %d", w.Code, tc.wantStatus)
			}
			if tc.name == "success" && !repo.notifs[0].Read {
				t.Error("notification should be marked read")
			}
		})
	}
}

// ---- mark-all-read tests ----

func TestHandler_MarkAllRead(t *testing.T) {
	repo := &mockNotifRepo{
		notifs: []Notification{
			{ID: uuid.New(), Type: TypePriceDrop, CreatedAt: time.Now()},
			{ID: uuid.New(), Type: TypeTargetHit, CreatedAt: time.Now()},
		},
	}
	req := httptest.NewRequest(http.MethodPost, "/read-all", nil)
	w := httptest.NewRecorder()

	h := NewHandler(repo)
	h.markAllRead(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d; want 204", w.Code)
	}
	for _, n := range repo.notifs {
		if !n.Read {
			t.Errorf("notification %s should be marked read", n.ID)
		}
	}
}

func TestHandler_MarkAllRead_RepoError(t *testing.T) {
	repo := &mockNotifRepo{markAllReadErr: errors.New("db error")}
	req := httptest.NewRequest(http.MethodPost, "/read-all", nil)
	w := httptest.NewRecorder()

	NewHandler(repo).markAllRead(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want 500", w.Code)
	}
}

// ---- unread-count tests ----

func TestHandler_UnreadCount(t *testing.T) {
	repo := &mockNotifRepo{
		notifs: []Notification{
			{ID: uuid.New(), Read: false},
			{ID: uuid.New(), Read: true},
			{ID: uuid.New(), Read: false},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/unread-count", nil)
	w := httptest.NewRecorder()

	NewHandler(repo).unreadCount(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", w.Code)
	}
	var got map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["count"] != 2 {
		t.Errorf("count = %d; want 2", got["count"])
	}
}

func TestHandler_UnreadCount_RepoError(t *testing.T) {
	repo := &mockNotifRepo{unreadCountErr: errors.New("db error")}
	req := httptest.NewRequest(http.MethodGet, "/unread-count", nil)
	w := httptest.NewRecorder()

	NewHandler(repo).unreadCount(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d; want 500", w.Code)
	}
}
