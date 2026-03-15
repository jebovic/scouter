package dealcalendar

import (
	"testing"
	"time"
)

func TestNewCalendar(t *testing.T) {
	cal := NewCalendar()
	if cal == nil {
		t.Fatal("NewCalendar returned nil")
	}

	events := cal.GetAll()
	if len(events) != 11 {
		t.Errorf("expected 11 events, got %d", len(events))
	}
}

func TestGetAll(t *testing.T) {
	cal := NewCalendar()
	events := cal.GetAll()

	if len(events) == 0 {
		t.Fatal("GetAll returned no events")
	}

	// Verify all events have required fields
	for i, event := range events {
		if event.Name == "" {
			t.Errorf("event %d has empty name", i)
		}
		if event.StartDate.IsZero() {
			t.Errorf("event %d has zero StartDate", i)
		}
		if event.EndDate.IsZero() {
			t.Errorf("event %d has zero EndDate", i)
		}
		if event.CategoryHint == "" {
			t.Errorf("event %d has empty CategoryHint", i)
		}
		if event.DiscountLevel == "" {
			t.Errorf("event %d has empty DiscountLevel", i)
		}
		if event.Description == "" {
			t.Errorf("event %d has empty Description", i)
		}
		if event.EndDate.Before(event.StartDate) {
			t.Errorf("event %d has EndDate before StartDate", i)
		}
	}
}

func TestGetUpcoming(t *testing.T) {
	cal := NewCalendar()

	// Set "now" to Jan 1, 2026
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	upcoming := cal.GetUpcoming(now)

	if len(upcoming) == 0 {
		t.Fatal("GetUpcoming returned no events for early 2026")
	}

	// Verify all returned events start on or after the given date
	for _, event := range upcoming {
		if event.StartDate.Before(now) {
			t.Errorf("event %s has StartDate %v before now %v", event.Name, event.StartDate, now)
		}
	}
}

func TestGetUpcomingAtYear2020(t *testing.T) {
	cal := NewCalendar()

	// Set "now" to Jan 1, 2020 — all 2026 events should be in the future
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	upcoming := cal.GetUpcoming(now)

	// All events should be returned
	allEvents := cal.GetAll()
	if len(upcoming) != len(allEvents) {
		t.Errorf("expected %d upcoming events, got %d", len(allEvents), len(upcoming))
	}
}

func TestGetUpcomingAtYear2030(t *testing.T) {
	cal := NewCalendar()

	// Set "now" to Jan 1, 2030 — no 2026 events should be in the future
	now := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	upcoming := cal.GetUpcoming(now)

	if len(upcoming) != 0 {
		t.Errorf("expected 0 upcoming events for 2030, got %d", len(upcoming))
	}
}

func TestFilterByCategory(t *testing.T) {
	cal := NewCalendar()
	allEvents := cal.GetAll()

	tests := []struct {
		category      string
		expectMatches int
		matchNames    []string
	}{
		{
			category:      "electronics",
			expectMatches: 6, // Amazon Prime Day, La Rentrée tech, Darty Days, Fnac Days, Black Friday, Cyber Monday (and events with "all")
			matchNames:    []string{"Amazon Prime Day", "La Rentrée tech", "Darty Days", "Fnac Days"},
		},
		{
			category:      "fashion",
			expectMatches: 3, // Soldes d'hiver, French Days printemps, Soldes d'été (plus all events)
			matchNames:    []string{},
		},
		{
			category:      "computing",
			expectMatches: 4, // Fnac Days, Cyber Monday, plus all events
			matchNames:    []string{"Fnac Days", "Cyber Monday"},
		},
	}

	for _, tt := range tests {
		filtered := cal.FilterByCategory(allEvents, tt.category)

		// Verify all returned events match the category or are "all"
		for _, event := range filtered {
			if event.CategoryHint != "all" && !contains(event.CategoryHint, tt.category) {
				t.Errorf("event %s does not match category %s", event.Name, tt.category)
			}
		}

		// Verify all expected names are present
		for _, expectedName := range tt.matchNames {
			found := false
			for _, event := range filtered {
				if event.Name == expectedName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("category %s: expected event %s not found in results", tt.category, expectedName)
			}
		}
	}
}

func TestFilterByEmptyCategory(t *testing.T) {
	cal := NewCalendar()
	allEvents := cal.GetAll()

	filtered := cal.FilterByCategory(allEvents, "")

	// Empty category should return all events
	if len(filtered) != len(allEvents) {
		t.Errorf("empty category: expected %d events, got %d", len(allEvents), len(filtered))
	}
}

func TestFilterByNonExistentCategory(t *testing.T) {
	cal := NewCalendar()
	allEvents := cal.GetAll()

	filtered := cal.FilterByCategory(allEvents, "nonexistent")

	// Non-existent category should still return "all" events
	for _, event := range filtered {
		if event.CategoryHint != "all" {
			t.Errorf("non-existent category: found event %s with categoryHint %s", event.Name, event.CategoryHint)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		csv      string
		value    string
		expected bool
	}{
		{
			csv:      "electronics,computing",
			value:    "electronics",
			expected: true,
		},
		{
			csv:      "electronics,computing",
			value:    "computing",
			expected: true,
		},
		{
			csv:      "electronics,computing",
			value:    "fashion",
			expected: false,
		},
		{
			csv:      "all",
			value:    "electronics",
			expected: false,
		},
	}

	for _, tt := range tests {
		result := contains(tt.csv, tt.value)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", tt.csv, tt.value, result, tt.expected)
		}
	}
}

func TestEventDatesAreValid(t *testing.T) {
	cal := NewCalendar()
	events := cal.GetAll()

	for _, event := range events {
		if event.StartDate.After(event.EndDate) {
			t.Errorf("event %s: StartDate is after EndDate", event.Name)
		}
		if event.StartDate.Year() != 2026 {
			t.Errorf("event %s: StartDate year is %d, expected 2026", event.Name, event.StartDate.Year())
		}
		if event.EndDate.Year() != 2026 {
			t.Errorf("event %s: EndDate year is %d, expected 2026", event.Name, event.EndDate.Year())
		}
	}
}

func TestEventDiscountLevels(t *testing.T) {
	cal := NewCalendar()
	events := cal.GetAll()

	validLevels := map[string]bool{"low": true, "medium": true, "high": true}

	for _, event := range events {
		if !validLevels[event.DiscountLevel] {
			t.Errorf("event %s has invalid discountLevel %q", event.Name, event.DiscountLevel)
		}
	}
}

func TestEventCategories(t *testing.T) {
	cal := NewCalendar()
	events := cal.GetAll()

	validCats := map[string]bool{"all": true, "electronics": true, "fashion": true, "travel": true, "computing": true}

	for _, event := range events {
		// categoryHint can be comma-separated
		fields := split(event.CategoryHint, ",")
		for _, field := range fields {
			if !validCats[field] {
				t.Errorf("event %s has invalid category %q in hint %q", event.Name, field, event.CategoryHint)
			}
		}
	}
}
