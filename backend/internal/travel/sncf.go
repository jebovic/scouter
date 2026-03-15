package travel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// SNCFClient fetches train journey data from the SNCF/Navitia API.
type SNCFClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewSNCFClient returns a client configured with apiKey and baseURL.
// When apiKey is empty, SearchTrains returns an empty slice immediately.
func NewSNCFClient(apiKey, baseURL string) *SNCFClient {
	return &SNCFClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// navitiaResponse mirrors the subset of the Navitia/SNCF JSON response
// required to build TrainOffer values.
type navitiaResponse struct {
	Journeys []struct {
		Duration            int    `json:"duration"` // seconds
		DepartureDatetime   string `json:"departure_date_time"`
		ArrivalDatetime     string `json:"arrival_date_time"`
		Sections            []struct {
			Type                string `json:"type"`
			DisplayInformations struct {
				CommercialMode string `json:"commercial_mode"`
				Name           string `json:"name"`
			} `json:"display_informations"`
		} `json:"sections"`
	} `json:"journeys"`
}

// navitiaTimeLayout is the datetime format used by the Navitia API.
const navitiaTimeLayout = "20060102T150405"

// SearchTrains queries the SNCF/Navitia API for train journeys between from
// and to on date (format "YYYY-MM-DD"). An empty apiKey causes the method to
// return immediately with an empty slice and no error.
func (c *SNCFClient) SearchTrains(ctx context.Context, from, to, date string) ([]TrainOffer, error) {
	if c.apiKey == "" {
		return []TrainOffer{}, nil
	}

	// Parse date and build Navitia datetime (YYYYMMDDTHHMMSS at 08:00).
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("sncf: invalid date %q: %w", date, err)
	}
	navDatetime := d.Format("20060102") + "T080000"

	endpoint := fmt.Sprintf("%s/journeys", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("sncf: build request: %w", err)
	}

	q := url.Values{}
	q.Set("from", from)
	q.Set("to", to)
	q.Set("datetime", navDatetime)
	q.Set("data_freshness", "realtime")
	req.URL.RawQuery = q.Encode()

	// SNCF uses HTTP Basic auth: API key as username, empty password.
	req.SetBasicAuth(c.apiKey, "")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sncf: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sncf: unexpected status %d", resp.StatusCode)
	}

	var nr navitiaResponse
	if err := json.NewDecoder(resp.Body).Decode(&nr); err != nil {
		return nil, fmt.Errorf("sncf: decode response: %w", err)
	}

	offers := make([]TrainOffer, 0, len(nr.Journeys))
	for _, j := range nr.Journeys {
		departAt, _ := time.ParseInLocation(navitiaTimeLayout, j.DepartureDatetime, time.UTC)
		arriveAt, _ := time.ParseInLocation(navitiaTimeLayout, j.ArrivalDatetime, time.UTC)
		durationMin := j.Duration / 60

		// Extract operator and train number from the first public_transport section.
		var operator, trainNum string
		for _, sec := range j.Sections {
			if sec.Type == "public_transport" {
				operator = sec.DisplayInformations.CommercialMode
				trainNum = sec.DisplayInformations.Name
				break
			}
		}

		offers = append(offers, TrainOffer{
			Operator:  operator,
			TrainNum:  trainNum,
			Departure: from,
			Arrival:   to,
			DepartAt:  departAt,
			ArriveAt:  arriveAt,
			Duration:  durationMin,
			PriceEUR:  0, // not available in free SNCF API
			Source:    "sncf",
		})
	}
	return offers, nil
}
