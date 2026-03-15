package travel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AviationStackClient fetches real-time flight data from the Aviationstack API.
type AviationStackClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewAviationStackClient returns a client configured with apiKey and baseURL.
// When apiKey is empty, SearchFlights returns an empty slice immediately
// (graceful degradation for environments without an API key).
func NewAviationStackClient(apiKey, baseURL string) *AviationStackClient {
	return &AviationStackClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// aviationResponse mirrors the subset of the Aviationstack JSON response
// required to build FlightOffer values.
type aviationResponse struct {
	Data []struct {
		Airline struct {
			Name string `json:"name"`
		} `json:"airline"`
		Flight struct {
			IATA string `json:"iata"`
		} `json:"flight"`
		Departure struct {
			IATA      string `json:"iata"`
			Scheduled string `json:"scheduled"`
		} `json:"departure"`
		Arrival struct {
			IATA      string `json:"iata"`
			Scheduled string `json:"scheduled"`
		} `json:"arrival"`
	} `json:"data"`
}

// SearchFlights queries Aviationstack for flights between from and to on date
// (format "YYYY-MM-DD"). An empty apiKey causes the method to return
// immediately with an empty slice and no error.
func (c *AviationStackClient) SearchFlights(ctx context.Context, from, to, date string) ([]FlightOffer, error) {
	if c.apiKey == "" {
		return []FlightOffer{}, nil
	}

	url := fmt.Sprintf(
		"%s/v1/flights?access_key=%s&dep_iata=%s&arr_iata=%s&flight_date=%s",
		c.baseURL, c.apiKey, from, to, date,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("aviationstack: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("aviationstack: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("aviationstack: unexpected status %d", resp.StatusCode)
	}

	var ar aviationResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, fmt.Errorf("aviationstack: decode response: %w", err)
	}

	offers := make([]FlightOffer, 0, len(ar.Data))
	for _, d := range ar.Data {
		departAt, _ := time.Parse(time.RFC3339, d.Departure.Scheduled)
		arriveAt, _ := time.Parse(time.RFC3339, d.Arrival.Scheduled)

		offers = append(offers, FlightOffer{
			Airline:   d.Airline.Name,
			FlightNum: d.Flight.IATA,
			Departure: d.Departure.IATA,
			Arrival:   d.Arrival.IATA,
			DepartAt:  departAt,
			ArriveAt:  arriveAt,
			PriceEUR:  0, // not available in free tier
			Currency:  "EUR",
			Source:    "aviationstack",
		})
	}
	return offers, nil
}
