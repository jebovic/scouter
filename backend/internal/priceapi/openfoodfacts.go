package priceapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var offSupportedCategories = map[string]bool{
	"food": true, "grocery": true, "cooking": true, "custom": true,
}

// OpenFoodFactsProvider searches the Open Food Facts database for food products
// and enriches results with real prices from the Open Prices API.
type OpenFoodFactsProvider struct {
	client        *http.Client
	baseURL       string // default "https://world.openfoodfacts.org"
	pricesBaseURL string // default "https://prices.openfoodfacts.org"
}

// NewOpenFoodFactsProvider creates a new OpenFoodFactsProvider using the given HTTP client.
func NewOpenFoodFactsProvider(client *http.Client) *OpenFoodFactsProvider {
	return &OpenFoodFactsProvider{
		client:        client,
		baseURL:       "https://world.openfoodfacts.org",
		pricesBaseURL: "https://prices.openfoodfacts.org",
	}
}

// Name returns the provider identifier.
func (p *OpenFoodFactsProvider) Name() string { return "OpenFoodFacts" }

// Supports returns true for food-related categories.
func (p *OpenFoodFactsProvider) Supports(category string) bool {
	return offSupportedCategories[category]
}

// Search queries Open Food Facts for matching products and enriches the first
// three (with non-empty codes) with real price data from the Prices API.
func (p *OpenFoodFactsProvider) Search(ctx context.Context, q ProviderQuery) ([]PriceListing, error) {
	searchBase := p.baseURL
	if strings.HasPrefix(q.Locale, "fr") {
		searchBase = strings.Replace(searchBase, "world.openfoodfacts.org", "fr.openfoodfacts.org", 1)
	}

	maxResults := q.MaxResults
	if maxResults <= 0 {
		maxResults = 10
	}

	searchURL := fmt.Sprintf(
		"%s/cgi/search.pl?search_terms=%s&search_simple=1&action=process&json=1&page_size=%d&fields=code,product_name,brands,image_front_small_url,stores",
		searchBase,
		url.QueryEscape(q.ProductName),
		maxResults,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build search request: %w", err)
	}
	req.Header.Set("User-Agent", "Scouter/1.0 (personal-spending-tool)")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	var searchResp offSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	listings := make([]PriceListing, 0, len(searchResp.Products))
	priceCount := 0

	for _, prod := range searchResp.Products {
		var price float64
		var currency string

		// Fetch price for first 3 products with valid codes.
		if prod.Code != "" && priceCount < 3 {
			priceCount++
			price, currency = p.fetchPrice(ctx, prod.Code)
		}
		if currency == "" {
			currency = "EUR"
		}

		listings = append(listings, PriceListing{
			Source:      p.Name(),
			ProductName: prod.ProductName,
			Brand:       prod.Brands,
			Price:       price,
			Currency:    currency,
			ImageURL:    prod.ImageFrontSmallURL,
			Stores:      prod.Stores,
			Condition:   "new",
			UpdatedAt:   time.Now().UTC(),
		})
	}

	return listings, nil
}

// fetchPrice retrieves the most recent price for the given product code from the
// Open Prices API. Returns (0, "") if no price data is found or the request fails.
func (p *OpenFoodFactsProvider) fetchPrice(ctx context.Context, code string) (float64, string) {
	pricesURL := fmt.Sprintf(
		"%s/api/v1/prices?product_code=%s&order_by=-date&page_size=1",
		p.pricesBaseURL,
		url.QueryEscape(code),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pricesURL, nil)
	if err != nil {
		return 0, ""
	}
	req.Header.Set("User-Agent", "Scouter/1.0 (personal-spending-tool)")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, ""
	}

	var pr offPricesResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return 0, ""
	}

	if len(pr.Items) == 0 {
		return 0, ""
	}
	return pr.Items[0].Price, pr.Items[0].Currency
}

// Internal structs for JSON parsing.

type offSearchResponse struct {
	Count    int          `json:"count"`
	Products []offProduct `json:"products"`
}

type offProduct struct {
	Code               string `json:"code"`
	ProductName        string `json:"product_name"`
	Brands             string `json:"brands"`
	ImageFrontSmallURL string `json:"image_front_small_url"`
	Stores             string `json:"stores"`
}

type offPricesResponse struct {
	Items []struct {
		Price    float64 `json:"price"`
		Currency string  `json:"currency"`
		Date     string  `json:"date"`
	} `json:"items"`
}
