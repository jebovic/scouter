package priceapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// offTestServer builds a test server that handles both the search and prices APIs.
// products: list of products to return from search.
// priceByCode: map of product_code → price (omitted = no price entry).
func offTestServer(t *testing.T, products []offProduct, priceByCode map[string]float64, searchErr bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prices API
		if strings.HasPrefix(r.URL.Path, "/api/v1/prices") {
			code := r.URL.Query().Get("product_code")
			price, ok := priceByCode[code]
			if !ok {
				_ = json.NewEncoder(w).Encode(offPricesResponse{Items: nil})
				return
			}
			resp := offPricesResponse{
				Items: []struct {
					Price    float64 `json:"price"`
					Currency string  `json:"currency"`
					Date     string  `json:"date"`
				}{
					{Price: price, Currency: "EUR", Date: "2026-01-01"},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}

		// Search API
		if searchErr {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := offSearchResponse{
			Count:    len(products),
			Products: products,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestOpenFoodFacts_HappyPath(t *testing.T) {
	products := []offProduct{
		{Code: "111", ProductName: "Cheese", Brands: "BrandA", ImageFrontSmallURL: "http://img/1.jpg", Stores: "Carrefour"},
		{Code: "222", ProductName: "Milk", Brands: "BrandB", ImageFrontSmallURL: "http://img/2.jpg", Stores: "Leclerc"},
		{Code: "333", ProductName: "Butter", Brands: "BrandC", ImageFrontSmallURL: "http://img/3.jpg", Stores: ""},
	}
	prices := map[string]float64{"111": 3.50, "222": 1.20}

	srv := offTestServer(t, products, prices, false)
	defer srv.Close()

	p := NewOpenFoodFactsProvider(srv.Client())
	// Override base URL and prices URL to point at test server.
	p.baseURL = srv.URL
	p.pricesBaseURL = srv.URL

	listings, err := p.Search(context.Background(), ProviderQuery{
		ProductName: "cheese",
		Category:    "food",
		Locale:      "en",
		MaxResults:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(listings) != 3 {
		t.Fatalf("expected 3 listings, got %d", len(listings))
	}

	// Cheese: price=3.50
	if listings[0].Price != 3.50 {
		t.Errorf("listing[0] price: want 3.50, got %f", listings[0].Price)
	}
	// Milk: price=1.20
	if listings[1].Price != 1.20 {
		t.Errorf("listing[1] price: want 1.20, got %f", listings[1].Price)
	}
	// Butter: no price
	if listings[2].Price != 0 {
		t.Errorf("listing[2] price: want 0, got %f", listings[2].Price)
	}
	if listings[0].Currency != "EUR" {
		t.Errorf("expected EUR currency, got %q", listings[0].Currency)
	}
	if listings[0].Source != "OpenFoodFacts" {
		t.Errorf("expected source=OpenFoodFacts, got %q", listings[0].Source)
	}
}

func TestOpenFoodFacts_EmptyResults(t *testing.T) {
	srv := offTestServer(t, nil, nil, false)
	defer srv.Close()

	p := NewOpenFoodFactsProvider(srv.Client())
	p.baseURL = srv.URL
	p.pricesBaseURL = srv.URL

	listings, err := p.Search(context.Background(), ProviderQuery{
		ProductName: "nonexistent",
		Category:    "food",
		MaxResults:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(listings) != 0 {
		t.Errorf("expected 0 listings, got %d", len(listings))
	}
}

func TestOpenFoodFacts_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{invalid json"))
	}))
	defer srv.Close()

	p := NewOpenFoodFactsProvider(srv.Client())
	p.baseURL = srv.URL
	p.pricesBaseURL = srv.URL

	_, err := p.Search(context.Background(), ProviderQuery{
		ProductName: "cheese",
		Category:    "food",
		MaxResults:  5,
	})
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestOpenFoodFacts_PricesAPIFails(t *testing.T) {
	products := []offProduct{
		{Code: "111", ProductName: "Cheese", Brands: "BrandA"},
		{Code: "222", ProductName: "Milk", Brands: "BrandB"},
	}

	pricesSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer pricesSrv.Close()

	searchSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := offSearchResponse{Count: 2, Products: products}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer searchSrv.Close()

	p := NewOpenFoodFactsProvider(searchSrv.Client())
	p.baseURL = searchSrv.URL
	p.pricesBaseURL = pricesSrv.URL

	listings, err := p.Search(context.Background(), ProviderQuery{
		ProductName: "cheese",
		Category:    "food",
		MaxResults:  5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Prices API failure → all prices should be 0.
	for _, l := range listings {
		if l.Price != 0 {
			t.Errorf("expected price=0 when prices API fails, got %f for %q", l.Price, l.ProductName)
		}
	}
}

func TestOpenFoodFacts_Supports(t *testing.T) {
	p := NewOpenFoodFactsProvider(http.DefaultClient)
	supported := []string{"food", "grocery", "cooking", "custom"}
	for _, cat := range supported {
		if !p.Supports(cat) {
			t.Errorf("expected Supports(%q)=true, got false", cat)
		}
	}
	notSupported := []string{"electronics", "travel", "renovation", "computing", ""}
	for _, cat := range notSupported {
		if p.Supports(cat) {
			t.Errorf("expected Supports(%q)=false, got true", cat)
		}
	}
}

func TestOpenFoodFacts_FrenchLocale(t *testing.T) {
	var requestedHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedHost = r.Host
		if strings.HasPrefix(r.URL.Path, "/api/v1/prices") {
			_ = json.NewEncoder(w).Encode(offPricesResponse{})
			return
		}
		_ = json.NewEncoder(w).Encode(offSearchResponse{})
	}))
	defer srv.Close()

	p := NewOpenFoodFactsProvider(srv.Client())
	// We test locale-based host switching by capturing which URL was called.
	// Since httptest only gives one server, we use a custom transport to verify.
	var capturedURL string
	transport := &captureTransport{
		inner:   srv.Client().Transport,
		capture: &capturedURL,
	}
	p.client = &http.Client{Transport: transport}
	p.baseURL = "http://world.openfoodfacts.org"
	p.pricesBaseURL = srv.URL
	// Redirect world.openfoodfacts.org and fr.openfoodfacts.org to our test server.
	p.client = &http.Client{Transport: &rewriteTransport{
		base:    srv.URL,
		inner:   srv.Client().Transport,
		capture: &capturedURL,
	}}

	_, _ = p.Search(context.Background(), ProviderQuery{
		ProductName: "fromage",
		Category:    "food",
		Locale:      "fr_FR",
		MaxResults:  5,
	})

	_ = requestedHost // may be empty if srv not called directly
	if !strings.Contains(capturedURL, "fr.openfoodfacts.org") {
		t.Errorf("expected French locale to use fr.openfoodfacts.org, got URL: %q", capturedURL)
	}
}

// captureTransport records the last request URL and delegates to inner.
type captureTransport struct {
	inner   http.RoundTripper
	capture *string
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	*t.capture = req.URL.String()
	return t.inner.RoundTrip(req)
}

// rewriteTransport rewrites requests for *.openfoodfacts.org to the base test server URL,
// while capturing the original URL for assertion.
type rewriteTransport struct {
	base    string
	inner   http.RoundTripper
	capture *string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	orig := req.URL.String()
	if strings.Contains(orig, "openfoodfacts.org") {
		*t.capture = orig
		// Rewrite host to test server.
		rewritten := req.Clone(req.Context())
		base := t.base
		rewritten.URL.Scheme = "http"
		rewritten.URL.Host = strings.TrimPrefix(base, "http://")
		return t.inner.RoundTrip(rewritten)
	}
	return t.inner.RoundTrip(req)
}
