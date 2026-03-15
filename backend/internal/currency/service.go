package currency

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	ecbURL     = "https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml"
	cacheTTL   = 24 * time.Hour
	httpTimeout = 10 * time.Second
)

// ecb XML envelope structures.
type ecbEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Cube    ecbCube  `xml:"Cube"`
}

type ecbCube struct {
	Inner ecbCubeTime `xml:"Cube"`
}

type ecbCubeTime struct {
	Time  string       `xml:"time,attr"`
	Rates []ecbRate    `xml:"Cube"`
}

type ecbRate struct {
	Currency string  `xml:"currency,attr"`
	Rate     float64 `xml:"rate,attr"`
}

// Service fetches and caches ECB exchange rates.
type Service struct {
	mu        sync.RWMutex
	cached    *Rates
	fetchedAt time.Time
	client    *http.Client
}

// NewService returns a new currency Service.
func NewService() *Service {
	return &Service{
		client: &http.Client{Timeout: httpTimeout},
	}
}

// GetRates returns cached rates, refreshing if the cache has expired.
func (s *Service) GetRates() (*Rates, error) {
	s.mu.RLock()
	if s.cached != nil && time.Since(s.fetchedAt) < cacheTTL {
		rates := s.cached
		s.mu.RUnlock()
		return rates, nil
	}
	s.mu.RUnlock()

	return s.refresh()
}

// Convert converts an amount from one currency to another.
func (s *Service) Convert(amount float64, from, to string) (float64, error) {
	rates, err := s.GetRates()
	if err != nil {
		return 0, err
	}

	if from == to {
		return amount, nil
	}

	fromRate, ok := rates.Rates[from]
	if !ok {
		return 0, fmt.Errorf("unsupported currency: %s", from)
	}
	toRate, ok := rates.Rates[to]
	if !ok {
		return 0, fmt.Errorf("unsupported currency: %s", to)
	}

	// Convert through EUR: from -> EUR -> to
	inEUR := amount / fromRate
	result := inEUR * toRate
	return result, nil
}

func (s *Service) refresh() (*Rates, error) {
	resp, err := s.client.Get(ecbURL)
	if err != nil {
		return nil, fmt.Errorf("fetch ECB rates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ECB returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read ECB response: %w", err)
	}

	var envelope ecbEnvelope
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parse ECB XML: %w", err)
	}

	ratesMap := make(map[string]float64, len(envelope.Cube.Inner.Rates)+1)
	ratesMap["EUR"] = 1.0
	for _, r := range envelope.Cube.Inner.Rates {
		ratesMap[r.Currency] = r.Rate
	}

	rates := &Rates{
		Base:  "EUR",
		Date:  envelope.Cube.Inner.Time,
		Rates: ratesMap,
	}

	s.mu.Lock()
	s.cached = rates
	s.fetchedAt = time.Now()
	s.mu.Unlock()

	return rates, nil
}
