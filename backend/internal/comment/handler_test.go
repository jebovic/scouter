package comment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

// ---------------------------------------------------------------------------
// Stub repository
// ---------------------------------------------------------------------------

type stubRepo struct {
	comments []Comment
	created  *Comment
	listErr  error
	createErr error
	deleteErr error
}

func (s *stubRepo) List(_ context.Context, _ string, _ *string) ([]Comment, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	if s.comments == nil {
		return []Comment{}, nil
	}
	return s.comments, nil
}

func (s *stubRepo) Create(_ context.Context, _ string, req CreateCommentRequest) (Comment, error) {
	if s.createErr != nil {
		return Comment{}, s.createErr
	}
	if s.created != nil {
		return *s.created, nil
	}
	return Comment{
		ID:        "test-id",
		MissionID: "test-mission-id",
		OptionID:  req.OptionID,
		Author:    req.Author,
		Body:      req.Body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (s *stubRepo) Delete(_ context.Context, _ string) error {
	return s.deleteErr
}

// stubMissionService resolves slug→missionID for tests.
type stubMissionService struct {
	missionID string
	notFound  bool
}

func (s *stubMissionService) GetMissionIDBySlug(_ context.Context, _ string) (string, bool, error) {
	if s.notFound {
		return "", false, nil
	}
	return s.missionID, true, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func makeRequest(method, path string, body any) *http.Request {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func withChiParams(req *http.Request, params map[string]string) *http.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func newHandler(repo Repository, missionID string) *Handler {
	svc := &stubMissionService{missionID: missionID}
	return NewHandler(repo, svc)
}

func newHandlerNotFound(repo Repository) *Handler {
	svc := &stubMissionService{notFound: true}
	return NewHandler(repo, svc)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// 1. List empty — returns 200 with empty array.
func TestList_Empty(t *testing.T) {
	h := newHandler(&stubRepo{}, "mission-uuid-123")

	req := withChiParam(makeRequest("GET", "/", nil), "slug", "my-mission")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []Comment
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d comments", len(result))
	}
}

// 2. List with optionId filter — passes filter to repo.
func TestList_WithOptionIDFilter(t *testing.T) {
	optID := "option-uuid-456"
	repo := &stubRepo{
		comments: []Comment{
			{ID: "c1", MissionID: "m1", OptionID: &optID, Author: "Alice", Body: "Great option!"},
		},
	}
	h := newHandler(repo, "mission-uuid-123")

	req := withChiParam(makeRequest("GET", "/?optionId="+optID, nil), "slug", "my-mission")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var result []Comment
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 comment, got %d", len(result))
	}
}

// 3. Create success — 201 with comment.
func TestCreate_Success(t *testing.T) {
	h := newHandler(&stubRepo{}, "mission-uuid-123")

	body := map[string]string{
		"author": "Alice",
		"body":   "Looks good!",
	}
	req := withChiParam(makeRequest("POST", "/", body), "slug", "my-mission")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var result Comment
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.Body != "Looks good!" {
		t.Errorf("expected body 'Looks good!', got %q", result.Body)
	}
}

// 4. Create body too long — 422.
func TestCreate_BodyTooLong(t *testing.T) {
	h := newHandler(&stubRepo{}, "mission-uuid-123")

	longBody := make([]byte, 1001)
	for i := range longBody {
		longBody[i] = 'x'
	}
	body := map[string]string{
		"author": "Alice",
		"body":   string(longBody),
	}
	req := withChiParam(makeRequest("POST", "/", body), "slug", "my-mission")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

// 5. Create empty body — 422.
func TestCreate_EmptyBody(t *testing.T) {
	h := newHandler(&stubRepo{}, "mission-uuid-123")

	body := map[string]string{
		"author": "Alice",
		"body":   "",
	}
	req := withChiParam(makeRequest("POST", "/", body), "slug", "my-mission")
	rr := httptest.NewRecorder()
	h.Create(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", rr.Code)
	}
}

// 6. Delete success — 204.
func TestDelete_Success(t *testing.T) {
	h := newHandler(&stubRepo{}, "mission-uuid-123")

	req := withChiParams(makeRequest("DELETE", "/", nil), map[string]string{
		"slug":      "my-mission",
		"commentId": "some-comment-id",
	})
	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

// 7. Delete not found — 404.
func TestDelete_NotFound(t *testing.T) {
	repo := &stubRepo{deleteErr: pgx.ErrNoRows}
	h := newHandler(repo, "mission-uuid-123")

	req := withChiParams(makeRequest("DELETE", "/", nil), map[string]string{
		"slug":      "my-mission",
		"commentId": "nonexistent-id",
	})
	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

// 8. Invalid slug (mission not found) — 404.
func TestList_InvalidSlug_NotFound(t *testing.T) {
	h := newHandlerNotFound(&stubRepo{})

	req := withChiParam(makeRequest("GET", "/", nil), "slug", "nonexistent-slug")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown slug, got %d", rr.Code)
	}
}

