package shopping

import (
	"context"
	"encoding/csv"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// mockExportSnapshots stores test price snapshots for export tests (separate from price_history_test mocks).
var mockExportSnapshots = map[uuid.UUID][]PriceSnapshot{}

// exportMockRepo extends mockRepo to override ListPriceHistory with mockExportSnapshots.
type exportMockRepo struct {
	*mockRepo
}

func (m *exportMockRepo) ListPriceHistory(_ context.Context, itemID uuid.UUID) ([]PriceSnapshot, error) {
	if m.nextErr != nil {
		return nil, m.nextErr
	}
	return mockExportSnapshots[itemID], nil
}

// ---------------------------------------------------------------------------
// Handler tests for CSV export
// ---------------------------------------------------------------------------

func TestHandler_ExportPriceHistoryCSV_BasicExport(t *testing.T) {
	baseMock := newMockRepo()
	repo := &exportMockRepo{mockRepo: baseMock}
	svc := NewService(repo)
	h := NewHandler(svc)

	itemID := uuid.New()
	missionID := uuid.New()

	// Create test item
	item := &Item{
		ID:        itemID,
		MissionID: missionID,
		Name:      "Laptop Pro",
		Merchant:  "Amazon",
		Price:     1299.99,
	}
	repo.items[itemID] = item

	// Create test snapshots with French formatting requirements
	ts1 := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 3, 12, 10, 0, 0, 0, time.UTC)
	ts3 := time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)

	mockExportSnapshots[itemID] = []PriceSnapshot{
		{ID: uuid.New(), ItemID: itemID, Price: 1299.99, RecordedAt: ts1, Note: "manual"},
		{ID: uuid.New(), ItemID: itemID, Price: 1249.99, RecordedAt: ts2, Note: "scheduler"},
		{ID: uuid.New(), ItemID: itemID, Price: 1199.99, RecordedAt: ts3, Note: "deal"},
	}
	t.Cleanup(func() { delete(mockExportSnapshots, itemID) })

	router := chi.NewRouter()
	router.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/missions/"+missionID.String()+"/shopping/"+itemID.String()+"/price-history/export",
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}

	// Check Content-Type
	if !strings.Contains(w.Header().Get("Content-Type"), "text/csv") {
		t.Errorf("Content-Type = %s; want text/csv", w.Header().Get("Content-Type"))
	}

	// Check BOM
	body := w.Body.Bytes()
	if len(body) < 3 || body[0] != 0xef || body[1] != 0xbb || body[2] != 0xbf {
		t.Error("CSV missing UTF-8 BOM")
	}

	// Parse CSV (skip BOM)
	reader := csv.NewReader(strings.NewReader(string(body[3:])))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}

	// Check header
	if len(records) != 4 { // 1 header + 3 data rows
		t.Errorf("CSV rows = %d; want 4", len(records))
	}
	if len(records) > 0 {
		header := records[0]
		if len(header) != 5 || header[0] != "Date" || header[1] != "Prix" {
			t.Errorf("header = %v; want [Date Prix Devise Source Détaillant]", header)
		}
	}

	// Check first data row (French date format DD/MM/YYYY and comma decimal)
	if len(records) > 1 {
		row := records[1]
		if row[0] != "10/03/2026" {
			t.Errorf("first row date = %q; want 10/03/2026", row[0])
		}
		if row[1] != "1299,99" {
			t.Errorf("first row price = %q; want 1299,99", row[1])
		}
		if row[3] != "manual" {
			t.Errorf("first row source = %q; want manual", row[3])
		}
	}

	// Check Content-Disposition filename
	disposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "attachment") || !strings.Contains(disposition, "prix-") {
		t.Errorf("Content-Disposition = %q; want attachment with prix- prefix", disposition)
	}
}

func TestHandler_ExportPriceHistoryCSV_InvalidItemID(t *testing.T) {
	baseMock := newMockRepo()
	repo := &exportMockRepo{mockRepo: baseMock}
	svc := NewService(repo)
	h := NewHandler(svc)
	missionID := uuid.New()

	router := chi.NewRouter()
	router.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/missions/"+missionID.String()+"/shopping/not-a-uuid/price-history/export",
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d; want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_ExportPriceHistoryCSV_ItemNotFound(t *testing.T) {
	baseMock := newMockRepo()
	repo := &exportMockRepo{mockRepo: baseMock}
	svc := NewService(repo)
	h := NewHandler(svc)
	itemID := uuid.New()
	missionID := uuid.New()

	router := chi.NewRouter()
	router.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/missions/"+missionID.String()+"/shopping/"+itemID.String()+"/price-history/export",
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", w.Code, http.StatusNotFound)
	}
}

func TestHandler_ExportPriceHistoryCSV_EmptyHistory(t *testing.T) {
	baseMock := newMockRepo()
	repo := &exportMockRepo{mockRepo: baseMock}
	svc := NewService(repo)
	h := NewHandler(svc)
	itemID := uuid.New()
	missionID := uuid.New()

	// Create item with no price history
	item := &Item{
		ID:        itemID,
		MissionID: missionID,
		Name:      "Widget",
		Merchant:  "Store",
		Price:     99.99,
	}
	repo.items[itemID] = item
	mockExportSnapshots[itemID] = []PriceSnapshot{}
	t.Cleanup(func() { delete(mockExportSnapshots, itemID) })

	router := chi.NewRouter()
	router.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/missions/"+missionID.String()+"/shopping/"+itemID.String()+"/price-history/export",
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d; want %d", w.Code, http.StatusOK)
	}

	// Should have header only
	body := w.Body.Bytes()
	reader := csv.NewReader(strings.NewReader(string(body[3:]))) // skip BOM
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}

	if len(records) != 1 { // header only
		t.Errorf("CSV rows = %d; want 1 (header only)", len(records))
	}
}

func TestHandler_ExportPriceHistoryCSV_FilenameSanitization(t *testing.T) {
	baseMock := newMockRepo()
	repo := &exportMockRepo{mockRepo: baseMock}
	svc := NewService(repo)
	h := NewHandler(svc)
	itemID := uuid.New()
	missionID := uuid.New()

	// Item name with special characters that should be sanitized
	item := &Item{
		ID:        itemID,
		MissionID: missionID,
		Name:      "USB-C Cable (High-Speed)",
		Merchant:  "Best Buy",
		Price:     29.99,
	}
	repo.items[itemID] = item
	mockExportSnapshots[itemID] = []PriceSnapshot{
		{ID: uuid.New(), ItemID: itemID, Price: 29.99, RecordedAt: time.Now(), Note: ""},
	}
	t.Cleanup(func() { delete(mockExportSnapshots, itemID) })

	router := chi.NewRouter()
	router.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/missions/"+missionID.String()+"/shopping/"+itemID.String()+"/price-history/export",
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	disposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "prix-usb-c-cable") {
		t.Errorf("filename = %q; should contain sanitized item name", disposition)
	}
}

func TestHandler_ExportPriceHistoryCSV_AllColumns(t *testing.T) {
	baseMock := newMockRepo()
	repo := &exportMockRepo{mockRepo: baseMock}
	svc := NewService(repo)
	h := NewHandler(svc)
	itemID := uuid.New()
	missionID := uuid.New()

	item := &Item{
		ID:        itemID,
		MissionID: missionID,
		Name:      "Monitor",
		Merchant:  "Newegg",
		Price:     349.99,
	}
	repo.items[itemID] = item

	ts := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
	mockExportSnapshots[itemID] = []PriceSnapshot{
		{ID: uuid.New(), ItemID: itemID, Price: 349.99, RecordedAt: ts, Note: "web-scrape"},
	}
	t.Cleanup(func() { delete(mockExportSnapshots, itemID) })

	router := chi.NewRouter()
	router.Route("/missions/{missionID}/shopping", func(r chi.Router) {
		r.Mount("/", h.Routes())
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/missions/"+missionID.String()+"/shopping/"+itemID.String()+"/price-history/export",
		nil,
	)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", w.Code, http.StatusOK)
	}

	body := w.Body.Bytes()
	reader := csv.NewReader(strings.NewReader(string(body[3:]))) // skip BOM

	// Read all records
	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read CSV: %v", err)
		}
		records = append(records, record)
	}

	if len(records) < 2 {
		t.Fatalf("expected at least 2 records (header + data), got %d", len(records))
	}

	row := records[1]
	if len(row) != 5 {
		t.Errorf("row has %d columns; want 5", len(row))
	}

	// Verify all columns are present and correctly populated
	if row[0] != "15/03/2026" {
		t.Errorf("Date column = %q; want 15/03/2026", row[0])
	}
	if row[1] != "349,99" {
		t.Errorf("Prix column = %q; want 349,99", row[1])
	}
	if row[2] != "EUR" {
		t.Errorf("Devise column = %q; want EUR", row[2])
	}
	if row[3] != "web-scrape" {
		t.Errorf("Source column = %q; want web-scrape", row[3])
	}
	if row[4] != "Newegg" {
		t.Errorf("Détaillant column = %q; want Newegg", row[4])
	}
}
