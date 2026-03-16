// backend/internal/researchjob/handler_test.go
package researchjob_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/researchjob"
)

// stubMissionService returns a fixed mission.
type stubMissionService struct {
	m *mission.Mission
}

func (s *stubMissionService) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return s.m, nil
}

// newTestRouter wires a Handler with mock deps onto a chi router.
func newTestRouter(repo researchjob.Repository, ms *mission.Mission) chi.Router {
	svc := &stubMissionService{m: ms}
	// MockAgent never panics, returns 0 options
	agent := &MockAgent{result: nil}
	h := researchjob.NewHandler(repo, svc, agent)
	r := chi.NewRouter()
	r.Route("/api/missions/{missionID}", func(r chi.Router) {
		r.Mount("/research", h.Routes())
	})
	return r
}

func TestHandler_Trigger_202(t *testing.T) {
	missionID := uuid.New()
	ms := &mission.Mission{ID: missionID, Name: "Test Mission"}
	repo := &MockRepository{}
	router := newTestRouter(repo, ms)

	body := strings.NewReader(`{"feedback":"cheaper please"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID.String()+"/research", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("want 202, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["job_id"] == nil {
		t.Error("expected job_id in response")
	}
	if resp["status"] != "pending" {
		t.Errorf("want status pending, got %v", resp["status"])
	}
}

func TestHandler_Trigger_409_Concurrent(t *testing.T) {
	missionID := uuid.New()
	ms := &mission.Mission{ID: missionID}
	repo := &MockRepository{}
	// Pre-insert a pending job (HasActiveJob returns true)
	repo.Create(context.Background(), missionID, "")

	router := newTestRouter(repo, ms)
	req := httptest.NewRequest(http.MethodPost, "/api/missions/"+missionID.String()+"/research", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("want 409, got %d", w.Code)
	}
}

func TestHandler_List_200(t *testing.T) {
	missionID := uuid.New()
	ms := &mission.Mission{ID: missionID}
	repo := &MockRepository{}
	repo.Create(context.Background(), missionID, "feedback")

	router := newTestRouter(repo, ms)
	req := httptest.NewRequest(http.MethodGet, "/api/missions/"+missionID.String()+"/research/jobs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["jobs"] == nil {
		t.Error("expected jobs array in response")
	}
}
