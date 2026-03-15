package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ErrNotFound is returned when the barcode is not in the Open Food Facts database.
var ErrNotFound = errors.New("product not found")

const offBaseURL = "https://world.openfoodfacts.org"

// offProduct maps the Open Food Facts API product fields.
type offProduct struct {
	Name       string `json:"product_name"`
	Brands     string `json:"brands"`
	Categories string `json:"categories"`
	ImageURL   string `json:"image_url"`
	NutriScore string `json:"nutriscore_grade"`
	EcoScore   string `json:"ecoscore_grade"`
	Stores     string `json:"stores"`
	Quantity   string `json:"quantity"`
}

// offResponse is the top-level Open Food Facts API response envelope.
type offResponse struct {
	Status  int        `json:"status"`
	Product offProduct `json:"product"`
}

// Looker fetches product information from Open Food Facts by EAN/UPC barcode.
type Looker struct {
	httpClient *http.Client
	baseURL    string
}

// NewLooker returns a Looker configured to use the public Open Food Facts API.
func NewLooker() *Looker {
	return &Looker{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    offBaseURL,
	}
}

// LookupByBarcode queries Open Food Facts for the given EAN/UPC barcode.
// Returns ErrNotFound when the product is absent from the database.
func (l *Looker) LookupByBarcode(ctx context.Context, barcode string) (ProductInfo, error) {
	url := fmt.Sprintf("%s/api/v2/product/%s.json", l.baseURL, barcode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Scouter/1.0 (contact@scouter.app)")

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProductInfo{}, fmt.Errorf("unexpected status %d from Open Food Facts", resp.StatusCode)
	}

	var off offResponse
	if err := json.NewDecoder(resp.Body).Decode(&off); err != nil {
		return ProductInfo{}, fmt.Errorf("decode response: %w", err)
	}

	if off.Status != 1 {
		return ProductInfo{}, ErrNotFound
	}

	stores := parseStores(off.Product.Stores)

	return ProductInfo{
		Barcode:    barcode,
		Name:       off.Product.Name,
		Brand:      off.Product.Brands,
		Category:   off.Product.Categories,
		ImageURL:   off.Product.ImageURL,
		NutriScore: off.Product.NutriScore,
		EcoScore:   off.Product.EcoScore,
		Stores:     stores,
		Quantity:   off.Product.Quantity,
		Source:     "openfoodfacts",
	}, nil
}

// parseStores splits a comma-separated store string, trimming whitespace from each entry.
// Returns nil (not an empty slice) when the input is blank.
func parseStores(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	stores := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			stores = append(stores, trimmed)
		}
	}
	return stores
}
