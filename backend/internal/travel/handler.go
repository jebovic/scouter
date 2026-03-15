package travel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/httputil"
)

// FlightSearcher is the narrow interface used by Handler so tests can inject
// fakes without depending on *AviationStackClient directly.
type FlightSearcher interface {
	SearchFlights(ctx context.Context, from, to, date string) ([]FlightOffer, error)
}

// TrainSearcher is the narrow interface used by Handler for SNCF lookups.
type TrainSearcher interface {
	SearchTrains(ctx context.Context, from, to, date string) ([]TrainOffer, error)
}

// Handler exposes travel search endpoints over HTTP.
type Handler struct {
	flights FlightSearcher
	trains  TrainSearcher
	cache   *Cache
}

// NewHandler returns a Handler wired to the given clients and cache.
// Both *AviationStackClient and *SNCFClient satisfy their respective interfaces.
func NewHandler(flights FlightSearcher, trains TrainSearcher, cache *Cache) *Handler {
	return &Handler{flights: flights, trains: trains, cache: cache}
}

// Routes returns a chi sub-router with the travel endpoints registered.
// Mount with: r.Mount("/api/travel", travelHandler.Routes())
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/flights", h.searchFlights)
	r.Get("/trains", h.searchTrains)
	return r
}

// searchFlights handles GET /flights?from=CDG&to=NCE&date=2026-04-01
func (h *Handler) searchFlights(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	date := r.URL.Query().Get("date")

	if from == "" || to == "" || date == "" {
		httputil.WriteError(w, http.StatusBadRequest, "from, to, and date query parameters are required")
		return
	}

	cacheKey := fmt.Sprintf("flights:%s:%s:%s", from, to, date)
	if cached, ok := h.cache.Get(cacheKey); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	offers, err := h.flights.SearchFlights(r.Context(), from, to, date)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "flight search unavailable")
		return
	}

	h.cache.Set(cacheKey, offers)
	httputil.WriteJSON(w, http.StatusOK, offers)
}

// searchTrains handles GET /trains?from=Paris+Gare+de+Lyon&to=Lyon+Part-Dieu&date=2026-04-01
func (h *Handler) searchTrains(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	date := r.URL.Query().Get("date")

	if from == "" || to == "" || date == "" {
		httputil.WriteError(w, http.StatusBadRequest, "from, to, and date query parameters are required")
		return
	}

	cacheKey := fmt.Sprintf("trains:%s:%s:%s", from, to, date)
	if cached, ok := h.cache.Get(cacheKey); ok {
		httputil.WriteJSON(w, http.StatusOK, cached)
		return
	}

	offers, err := h.trains.SearchTrains(r.Context(), from, to, date)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "train search unavailable")
		return
	}

	h.cache.Set(cacheKey, offers)
	httputil.WriteJSON(w, http.StatusOK, offers)
}
