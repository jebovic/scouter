package option

import (
	"bytes"
	"context"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// TestExportCSV_EmptyOptions returns CSV header only when no options exist
func TestExportCSV_EmptyOptions(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	missionID := uuid.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set up chi router context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ExportCSV(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("ExportCSV status = %d; want %d", w.Code, http.StatusOK)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/csv; charset=utf-8" {
		t.Errorf("Content-Type = %q; want %q", contentType, "text/csv; charset=utf-8")
	}

	disposition := w.Header().Get("Content-Disposition")
	if disposition == "" {
		t.Error("Content-Disposition header missing")
	}

	// Check BOM is present
	body := w.Body.Bytes()
	if len(body) < 3 || body[0] != 0xef || body[1] != 0xbb || body[2] != 0xbf {
		t.Error("BOM prefix not found")
	}

	// Parse CSV and verify header
	reader := csv.NewReader(bytes.NewReader(body[3:]))
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV header: %v", err)
	}

	if len(header) == 0 {
		t.Error("CSV header is empty")
	}
}

// TestExportCSV_SingleOption returns correct row for a single option
func TestExportCSV_SingleOption(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	missionID := uuid.New()
	optionID := uuid.New()

	// Create a test option
	opt := Option{
		ID:        optionID,
		MissionID: missionID,
		Name:      "MacBook Pro",
		Badge:     "recommended",
		PriceRange: &PriceRange{
			Min:  999.99,
			Max:  1299.99,
			Best: 999.99,
		},
		Attributes: []Attribute{
			{
				Key:   "cpu",
				Label: "Processor",
				Value: "Apple M3 Pro",
				Type:  "text",
			},
			{
				Key:   "memory",
				Label: "Memory",
				Value: 16,
				Type:  "text",
			},
		},
		CreatedAt: time.Now(),
	}

	repo.options[optionID] = &opt

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set up chi router context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ExportCSV(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("ExportCSV status = %d; want %d", w.Code, http.StatusOK)
	}

	// Parse CSV and check row
	body := w.Body.Bytes()
	reader := csv.NewReader(bytes.NewReader(body[3:])) // Skip BOM
	header, _ := reader.Read()
	row, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read CSV row: %v", err)
	}
	_ = row // Mark row as used below

	// Find index of Name column
	nameIdx := -1
	for i, h := range header {
		if h == "Nom" {
			nameIdx = i
			break
		}
	}
	if nameIdx == -1 {
		t.Error("Nom column not found in header")
	} else if row[nameIdx] != "MacBook Pro" {
		t.Errorf("Name = %q; want %q", row[nameIdx], "MacBook Pro")
	}
}

// TestExportCSV_MultipleOptions returns correct rows for multiple options
func TestExportCSV_MultipleOptions(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	missionID := uuid.New()

	opts := []Option{
		{
			ID:        uuid.New(),
			MissionID: missionID,
			Name:      "Option 1",
			Badge:     "recommended",
			CreatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			MissionID: missionID,
			Name:      "Option 2",
			Badge:     "watch",
			CreatedAt: time.Now().Add(time.Hour),
		},
		{
			ID:        uuid.New(),
			MissionID: missionID,
			Name:      "Option 3",
			Badge:     "rejected",
			CreatedAt: time.Now().Add(2 * time.Hour),
		},
	}

	for i := range opts {
		repo.options[opts[i].ID] = &opts[i]
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set up chi router context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ExportCSV(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("ExportCSV status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.Bytes()
	reader := csv.NewReader(bytes.NewReader(body[3:])) // Skip BOM
	_, _ = reader.Read() // Skip header

	rowCount := 0
	for {
		_, err := reader.Read()
		if err != nil {
			break
		}
		rowCount++
	}

	if rowCount != 3 {
		t.Errorf("Row count = %d; want 3", rowCount)
	}
}

// TestExportCSV_FrenchStatusLabels verifies French status translations
func TestExportCSV_FrenchStatusLabels(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	missionID := uuid.New()

	for i, badge := range []string{"recommended", "alternative", "watch", "rejected"} {
		optID := uuid.New()
		opt := Option{
			ID:        optID,
			MissionID: missionID,
			Name:      "Test " + badge,
			Badge:     badge,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
		}
		repo.options[optID] = &opt
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set up chi router context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ExportCSV(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("ExportCSV status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.Bytes()
	reader := csv.NewReader(bytes.NewReader(body[3:])) // Skip BOM
	header, _ := reader.Read()

	// Find Status column
	statusIdx := -1
	for i, h := range header {
		if h == "Statut" {
			statusIdx = i
			break
		}
	}
	if statusIdx == -1 {
		t.Error("Statut column not found in header")
		return
	}

	// Check each row has correct French label
	seenLabels := make(map[string]bool)
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}
		seenLabels[row[statusIdx]] = true
	}

	expectedLabels := map[string]bool{
		"Recommandé":     true,
		"Alternative":    true,
		"À surveiller":   true,
		"Rejeté":         true,
	}

	for expected := range expectedLabels {
		if !seenLabels[expected] {
			t.Errorf("Expected label %q not found in CSV", expected)
		}
	}
}

// TestExportCSV_BOMPrefix verifies UTF-8 BOM is present
func TestExportCSV_BOMPrefix(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	missionID := uuid.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set up chi router context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ExportCSV(w, r)

	body := w.Body.Bytes()
	if len(body) < 3 {
		t.Error("Response too short to contain BOM")
		return
	}

	if body[0] != 0xef || body[1] != 0xbb || body[2] != 0xbf {
		t.Errorf("BOM = [%x %x %x]; want [ef bb bf]", body[0], body[1], body[2])
	}
}

// TestExportCSV_ContentDisposition verifies correct header format
func TestExportCSV_ContentDisposition(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	missionID := uuid.New()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	// Set up chi router context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID.String())
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	handler.ExportCSV(w, r)

	disposition := w.Header().Get("Content-Disposition")
	if disposition == "" {
		t.Error("Content-Disposition header missing")
		return
	}

	if !bytes.Contains([]byte(disposition), []byte("attachment")) {
		t.Errorf("Content-Disposition = %q; should contain 'attachment'", disposition)
	}

	if !bytes.Contains([]byte(disposition), []byte(".csv")) {
		t.Errorf("Content-Disposition = %q; should contain filename with .csv", disposition)
	}
}

