package travel_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/travel"
)

// flightSearcher and trainSearcher are narrow interfaces used by the handler
// so we can inject fakes in tests without touching the real HTTP clients.

type fakeFlightClient struct {
	offers []travel.FlightOffer
	err    error
	called int
}

func (f *fakeFlightClient) SearchFlights(_ context.Context, from, to, date string) ([]travel.FlightOffer, error) {
	f.called++
	return f.offers, f.err
}

type fakeTrainClient struct {
	offers []travel.TrainOffer
	err    error
	called int
}

func (f *fakeTrainClient) SearchTrains(_ context.Context, from, to, date string) ([]travel.TrainOffer, error) {
	f.called++
	return f.offers, f.err
}

// mountRouter returns a chi router with travel routes mounted under /api/travel.
func mountRouter(h *travel.Handler) http.Handler {
	r := chi.NewRouter()
	r.Mount("/api/travel", h.Routes())
	return r
}

// ── /api/travel/flights ──────────────────────────────────────────────────────

func TestHandler_Flights_HappyPath(t *testing.T) {
	fc := &fakeFlightClient{
		offers: []travel.FlightOffer{
			{
				Airline:   "Air France",
				FlightNum: "AF1234",
				Departure: "CDG",
				Arrival:   "NCE",
				DepartAt:  time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC),
				ArriveAt:  time.Date(2026, 4, 1, 9, 30, 0, 0, time.UTC),
				PriceEUR:  0,
				Currency:  "EUR",
				Source:    "aviationstack",
			},
		},
	}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got []travel.FlightOffer
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 offer, got %d", len(got))
	}
	if got[0].FlightNum != "AF1234" {
		t.Errorf("expected AF1234, got %s", got[0].FlightNum)
	}
	if fc.called != 1 {
		t.Errorf("expected client to be called once, got %d", fc.called)
	}
}

func TestHandler_Flights_CacheHit(t *testing.T) {
	fc := &fakeFlightClient{
		offers: []travel.FlightOffer{{FlightNum: "AF9999"}},
	}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)
	router := mountRouter(h)

	req1 := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE&date=2026-04-01", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — should be served from cache; client not called again.
	req2 := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE&date=2026-04-01", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if fc.called != 1 {
		t.Errorf("expected client called exactly once (cache hit on 2nd), got %d", fc.called)
	}
}

func TestHandler_Flights_MissingFrom(t *testing.T) {
	fc := &fakeFlightClient{}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/flights?to=NCE&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Flights_MissingTo(t *testing.T) {
	fc := &fakeFlightClient{}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Flights_MissingDate(t *testing.T) {
	fc := &fakeFlightClient{}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_Flights_ClientError(t *testing.T) {
	fc := &fakeFlightClient{err: context.DeadlineExceeded}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

func TestHandler_Flights_EmptyResult(t *testing.T) {
	fc := &fakeFlightClient{offers: []travel.FlightOffer{}}
	tc := &fakeTrainClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var got []travel.FlightOffer
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty array, got %d items", len(got))
	}
}

// ── /api/travel/trains ───────────────────────────────────────────────────────

func TestHandler_Trains_HappyPath(t *testing.T) {
	tc := &fakeTrainClient{
		offers: []travel.TrainOffer{
			{
				Operator:  "SNCF",
				TrainNum:  "TGV6201",
				Departure: "Paris Gare de Lyon",
				Arrival:   "Lyon Part-Dieu",
				DepartAt:  time.Date(2026, 4, 1, 7, 0, 0, 0, time.UTC),
				ArriveAt:  time.Date(2026, 4, 1, 8, 59, 0, 0, time.UTC),
				Duration:  119,
				PriceEUR:  0,
				Source:    "sncf",
			},
		},
	}
	fc := &fakeFlightClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/trains?from=Paris+Gare+de+Lyon&to=Lyon+Part-Dieu&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got []travel.TrainOffer
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 offer, got %d", len(got))
	}
	if got[0].TrainNum != "TGV6201" {
		t.Errorf("expected TGV6201, got %s", got[0].TrainNum)
	}
}

func TestHandler_Trains_CacheHit(t *testing.T) {
	tc := &fakeTrainClient{
		offers: []travel.TrainOffer{{TrainNum: "TGV0001"}},
	}
	fc := &fakeFlightClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)
	router := mountRouter(h)

	req1 := httptest.NewRequest(http.MethodGet, "/api/travel/trains?from=Paris&to=Lyon&date=2026-04-01", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request: %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/travel/trains?from=Paris&to=Lyon&date=2026-04-01", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request: %d", w2.Code)
	}

	if tc.called != 1 {
		t.Errorf("expected client called once (cache hit), got %d", tc.called)
	}
}

func TestHandler_Trains_MissingParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"missing from", "/api/travel/trains?to=Lyon&date=2026-04-01"},
		{"missing to", "/api/travel/trains?from=Paris&date=2026-04-01"},
		{"missing date", "/api/travel/trains?from=Paris&to=Lyon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &fakeFlightClient{}
			tc := &fakeTrainClient{}
			cache := travel.NewCache(time.Hour)
			h := travel.NewHandler(fc, tc, cache)

			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			mountRouter(h).ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestHandler_Trains_ClientError(t *testing.T) {
	tc := &fakeTrainClient{err: context.DeadlineExceeded}
	fc := &fakeFlightClient{}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/travel/trains?from=Paris&to=Lyon&date=2026-04-01", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d", w.Code)
	}
}

// ── cache key isolation ──────────────────────────────────────────────────────

func TestHandler_CacheKeys_FlightsAndTrainsAreSeparate(t *testing.T) {
	fc := &fakeFlightClient{offers: []travel.FlightOffer{{FlightNum: "X"}}}
	tc := &fakeTrainClient{offers: []travel.TrainOffer{{TrainNum: "Y"}}}
	cache := travel.NewCache(time.Hour)
	h := travel.NewHandler(fc, tc, cache)
	router := mountRouter(h)

	// Populate cache for flights.
	r1 := httptest.NewRequest(http.MethodGet, "/api/travel/flights?from=CDG&to=NCE&date=2026-04-01", nil)
	router.ServeHTTP(httptest.NewRecorder(), r1)

	// Trains with same from/to/date must not hit the flights cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/travel/trains?from=CDG&to=NCE&date=2026-04-01", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("trains request failed: %d", w2.Code)
	}

	if tc.called != 1 {
		t.Errorf("train client should have been called once, got %d", tc.called)
	}
}
