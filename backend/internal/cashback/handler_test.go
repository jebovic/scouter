package cashback_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/cashback"
)

func setupRouter(t *testing.T) (*cashback.Repository, http.Handler) {
	t.Helper()
	repo := cashback.NewRepository()
	h := cashback.NewHandler(repo)
	r := chi.NewRouter()
	r.Mount("/api/cashback", h.Routes())
	return repo, r
}

func postJSON(t *testing.T, handler http.Handler, path string, body any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func getJSON(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestCreate_Success(t *testing.T) {
	_, router := setupRouter(t)

	body := map[string]any{
		"itemName":    "MacBook Pro",
		"merchant":    "Amazon",
		"amount":      45.0,
		"percentRate": 5.0,
	}
	rr := postJSON(t, router, "/api/cashback/", body, map[string]string{
		"X-Collaborator-ID": "user-1",
	})

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}

	var entry cashback.Entry
	if err := json.NewDecoder(rr.Body).Decode(&entry); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if entry.ItemName != "MacBook Pro" {
		t.Errorf("expected itemName MacBook Pro, got %q", entry.ItemName)
	}
	if entry.Status != "pending" {
		t.Errorf("expected status pending, got %q", entry.Status)
	}
	if entry.UserRef != "user-1" {
		t.Errorf("expected userRef user-1, got %q", entry.UserRef)
	}
}

func TestCreate_MissingUserRef(t *testing.T) {
	_, router := setupRouter(t)

	body := map[string]any{"itemName": "TV", "merchant": "Fnac", "amount": 10.0}
	rr := postJSON(t, router, "/api/cashback/", body, nil)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreate_MissingItemName(t *testing.T) {
	_, router := setupRouter(t)

	body := map[string]any{"merchant": "Fnac", "amount": 10.0}
	rr := postJSON(t, router, "/api/cashback/", body, map[string]string{
		"X-Collaborator-ID": "user-1",
	})

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rr.Code)
	}
}

func TestCreate_UserRefFromQueryParam(t *testing.T) {
	_, router := setupRouter(t)

	body := map[string]any{"itemName": "Phone", "merchant": "Darty", "amount": 8.0}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/cashback/?userRef=user-qp", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestList(t *testing.T) {
	_, router := setupRouter(t)

	// Create two entries for user-1, one for user-2
	for _, body := range []map[string]any{
		{"itemName": "Item A", "merchant": "M1", "amount": 5.0},
		{"itemName": "Item B", "merchant": "M2", "amount": 3.0},
	} {
		postJSON(t, router, "/api/cashback/", body, map[string]string{"X-Collaborator-ID": "user-1"})
	}
	postJSON(t, router, "/api/cashback/", map[string]any{"itemName": "Item C", "merchant": "M3", "amount": 1.0}, map[string]string{"X-Collaborator-ID": "user-2"})

	rr := getJSON(t, router, "/api/cashback/?userRef=user-1")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var entries []cashback.Entry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for user-1, got %d", len(entries))
	}
}

func TestList_MissingUserRef(t *testing.T) {
	_, router := setupRouter(t)
	rr := getJSON(t, router, "/api/cashback/")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestUpdate(t *testing.T) {
	_, router := setupRouter(t)

	// Create
	createBody := map[string]any{"itemName": "TV", "merchant": "Fnac", "amount": 20.0}
	createRR := postJSON(t, router, "/api/cashback/", createBody, map[string]string{"X-Collaborator-ID": "user-1"})
	var created cashback.Entry
	json.NewDecoder(createRR.Body).Decode(&created)

	// Update
	now := time.Now().UTC()
	updateBody := map[string]any{"status": "paid", "paidAt": now, "notes": "received!"}
	b, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/cashback/"+created.ID.String(), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var updated cashback.Entry
	json.NewDecoder(rr.Body).Decode(&updated)
	if updated.Status != "paid" {
		t.Errorf("expected status paid, got %q", updated.Status)
	}
	if updated.Notes != "received!" {
		t.Errorf("expected notes 'received!', got %q", updated.Notes)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	_, router := setupRouter(t)

	updateBody := map[string]any{"status": "paid"}
	b, _ := json.Marshal(updateBody)
	req := httptest.NewRequest(http.MethodPut, "/api/cashback/00000000-0000-0000-0000-000000000001", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestDelete(t *testing.T) {
	_, router := setupRouter(t)

	createBody := map[string]any{"itemName": "Laptop", "merchant": "Cdiscount", "amount": 30.0}
	createRR := postJSON(t, router, "/api/cashback/", createBody, map[string]string{"X-Collaborator-ID": "user-1"})
	var created cashback.Entry
	json.NewDecoder(createRR.Body).Decode(&created)

	req := httptest.NewRequest(http.MethodDelete, "/api/cashback/"+created.ID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}

	// Second delete should 404
	req2 := httptest.NewRequest(http.MethodDelete, "/api/cashback/"+created.ID.String(), nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on second delete, got %d", rr2.Code)
	}
}

func TestSummary(t *testing.T) {
	_, router := setupRouter(t)

	entries := []map[string]any{
		{"itemName": "A", "merchant": "M", "amount": 10.0, "status": "pending"},
		{"itemName": "B", "merchant": "M", "amount": 20.0, "status": "confirmed"},
		{"itemName": "C", "merchant": "M", "amount": 30.0, "status": "paid"},
	}
	for _, e := range entries {
		postJSON(t, router, "/api/cashback/", e, map[string]string{"X-Collaborator-ID": "user-s"})
	}

	rr := getJSON(t, router, "/api/cashback/summary?userRef=user-s")
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var s cashback.Summary
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if s.EntryCount != 3 {
		t.Errorf("expected 3 entries, got %d", s.EntryCount)
	}
	if s.TotalPending != 10.0 {
		t.Errorf("expected pending 10, got %f", s.TotalPending)
	}
	if s.TotalConfirmed != 20.0 {
		t.Errorf("expected confirmed 20, got %f", s.TotalConfirmed)
	}
	if s.TotalPaid != 30.0 {
		t.Errorf("expected paid 30, got %f", s.TotalPaid)
	}
}

func TestSummary_MissingUserRef(t *testing.T) {
	_, router := setupRouter(t)
	rr := getJSON(t, router, "/api/cashback/summary")
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
