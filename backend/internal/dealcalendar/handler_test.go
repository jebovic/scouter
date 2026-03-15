package dealcalendar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDealCalendarList(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		name               string
		queryParams        string
		expectStatus       int
		expectEventCount   int
		expectEventNames   []string
		checkCategoryFilter bool
		expectedCategories []string
	}{
		{
			name:             "GET /api/deal-calendar returns all events",
			queryParams:      "",
			expectStatus:     http.StatusOK,
			expectEventCount: 11,
		},
		{
			name:               "GET /api/deal-calendar?category=electronics filters correctly",
			queryParams:        "?category=electronics",
			expectStatus:       http.StatusOK,
			checkCategoryFilter: true,
			expectedCategories: []string{"electronics", "all"},
		},
		{
			name:               "GET /api/deal-calendar?category=computing filters correctly",
			queryParams:        "?category=computing",
			expectStatus:       http.StatusOK,
			checkCategoryFilter: true,
			expectedCategories: []string{"computing", "all"},
		},
		{
			name:             "GET /api/deal-calendar with empty category returns all",
			queryParams:      "?category=",
			expectStatus:     http.StatusOK,
			expectEventCount: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/deal-calendar"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.List(w, req)

			if w.Code != tt.expectStatus {
				t.Fatalf("expected status %d, got %d", tt.expectStatus, w.Code)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			events, ok := resp["events"].([]interface{})
			if !ok {
				t.Fatalf("events field is not an array")
			}

			if tt.expectEventCount > 0 && len(events) != tt.expectEventCount {
				t.Errorf("expected %d events, got %d", tt.expectEventCount, len(events))
			}

			if tt.checkCategoryFilter {
				for _, eventIface := range events {
					eventMap := eventIface.(map[string]interface{})
					categoryHint := eventMap["categoryHint"].(string)
					found := false
					for _, expectedCat := range tt.expectedCategories {
						if categoryHint == expectedCat || contains(categoryHint, tt.expectedCategories[0]) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("event %v has categoryHint %s, not in %v", eventMap["name"], categoryHint, tt.expectedCategories)
					}
				}
			}
		})
	}
}

func TestDealCalendarUpcomingFilter(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/api/deal-calendar?upcoming=true", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	events, ok := resp["events"].([]interface{})
	if !ok {
		t.Fatalf("events field is not an array")
	}

	// Verify all events are in the future
	now := time.Now()
	for _, eventIface := range events {
		eventMap := eventIface.(map[string]interface{})
		startDateStr := eventMap["startDate"].(string)
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			t.Fatalf("failed to parse startDate: %v", err)
		}
		if startDate.Before(now) {
			t.Errorf("event %v has startDate in the past: %v", eventMap["name"], startDate)
		}
	}
}

func TestDealCalendarCount(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/api/deal-calendar", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	count, ok := resp["count"].(float64)
	if !ok {
		t.Fatalf("count field is not a number")
	}

	events, ok := resp["events"].([]interface{})
	if !ok {
		t.Fatalf("events field is not an array")
	}

	if int(count) != len(events) {
		t.Errorf("count mismatch: count=%d, events=%d", int(count), len(events))
	}
}

func TestDealEventStructure(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/api/deal-calendar", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	events, ok := resp["events"].([]interface{})
	if !ok || len(events) == 0 {
		t.Fatalf("no events in response")
	}

	// Check first event has all required fields
	event := events[0].(map[string]interface{})

	requiredFields := []string{"name", "startDate", "endDate", "categoryHint", "discountLevel", "description"}
	for _, field := range requiredFields {
		if _, exists := event[field]; !exists {
			t.Errorf("event missing required field: %s", field)
		}
	}

	// Validate discountLevel is one of the expected values
	discountLevel := event["discountLevel"].(string)
	validLevels := []string{"low", "medium", "high"}
	found := false
	for _, level := range validLevels {
		if discountLevel == level {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("invalid discountLevel: %s", discountLevel)
	}
}

func TestDealCalendarResponseFormat(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/api/deal-calendar", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if _, hasEvents := resp["events"]; !hasEvents {
		t.Fatalf("response missing events field")
	}
	if _, hasCount := resp["count"]; !hasCount {
		t.Fatalf("response missing count field")
	}
}

func TestDealCalendarSorted(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest("GET", "/api/deal-calendar", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	events, _ := resp["events"].([]interface{})

	// Check that events are sorted by startDate
	var prevStart time.Time
	for i, eventIface := range events {
		eventMap := eventIface.(map[string]interface{})
		startDateStr := eventMap["startDate"].(string)
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			t.Fatalf("failed to parse startDate: %v", err)
		}

		if i > 0 && startDate.Before(prevStart) {
			t.Errorf("events not sorted by startDate at index %d", i)
		}
		prevStart = startDate
	}
}

func TestDealCalendarCombinedFilters(t *testing.T) {
	handler := NewHandler()

	// Test combining both category and upcoming filters
	req := httptest.NewRequest("GET", "/api/deal-calendar?category=electronics&upcoming=true", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	events, ok := resp["events"].([]interface{})
	if !ok {
		t.Fatalf("events field is not an array")
	}

	// Verify all returned events match both criteria:
	// 1. Category is electronics or all
	// 2. StartDate is in the future
	now := time.Now()
	for _, eventIface := range events {
		eventMap := eventIface.(map[string]interface{})

		// Check category
		categoryHint := eventMap["categoryHint"].(string)
		if categoryHint != "all" && !contains(categoryHint, "electronics") {
			t.Errorf("event %v has categoryHint %s, not electronics or all", eventMap["name"], categoryHint)
		}

		// Check upcoming (future date)
		startDateStr := eventMap["startDate"].(string)
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			t.Fatalf("failed to parse startDate: %v", err)
		}
		if startDate.Before(now) {
			t.Errorf("event %v has startDate in the past", eventMap["name"])
		}
	}
}
