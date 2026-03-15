package product

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLookupByBarcode_Success(t *testing.T) {
	response := offResponse{
		Status: 1,
		Product: offProduct{
			Name:        "Nutella",
			Brands:      "Ferrero",
			Categories:  "Spreads, Hazelnut spreads",
			ImageURL:    "https://images.openfoodfacts.org/nutella.jpg",
			NutriScore:  "e",
			EcoScore:    "c",
			Stores:      "Carrefour, Leclerc, Monoprix",
			Quantity:    "400g",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v2/product/3017620422003.json", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	info, err := looker.LookupByBarcode(context.Background(), "3017620422003")
	require.NoError(t, err)

	assert.Equal(t, "3017620422003", info.Barcode)
	assert.Equal(t, "Nutella", info.Name)
	assert.Equal(t, "Ferrero", info.Brand)
	assert.Equal(t, "Spreads, Hazelnut spreads", info.Category)
	assert.Equal(t, "https://images.openfoodfacts.org/nutella.jpg", info.ImageURL)
	assert.Equal(t, "e", info.NutriScore)
	assert.Equal(t, "c", info.EcoScore)
	assert.Equal(t, []string{"Carrefour", "Leclerc", "Monoprix"}, info.Stores)
	assert.Equal(t, "400g", info.Quantity)
	assert.Equal(t, "openfoodfacts", info.Source)
}

func TestLookupByBarcode_NotFound(t *testing.T) {
	response := offResponse{Status: 0}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := looker.LookupByBarcode(context.Background(), "0000000000000")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestLookupByBarcode_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := looker.LookupByBarcode(context.Background(), "1234567890123")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNotFound)
}

func TestLookupByBarcode_EmptyStores(t *testing.T) {
	response := offResponse{
		Status: 1,
		Product: offProduct{
			Name:   "Test Product",
			Brands: "Test Brand",
			Stores: "",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	info, err := looker.LookupByBarcode(context.Background(), "9876543210987")
	require.NoError(t, err)
	assert.Empty(t, info.Stores)
}

func TestLookupByBarcode_SingleStore(t *testing.T) {
	response := offResponse{
		Status: 1,
		Product: offProduct{
			Name:   "Test Product",
			Brands: "Test Brand",
			Stores: "Carrefour",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	info, err := looker.LookupByBarcode(context.Background(), "1111111111111")
	require.NoError(t, err)
	assert.Equal(t, []string{"Carrefour"}, info.Stores)
}

func TestLookupByBarcode_StoresWithSpaces(t *testing.T) {
	response := offResponse{
		Status: 1,
		Product: offProduct{
			Name:   "Test Product",
			Brands: "Test Brand",
			Stores: "  Carrefour , Leclerc  , Monoprix  ",
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	info, err := looker.LookupByBarcode(context.Background(), "2222222222222")
	require.NoError(t, err)
	assert.Equal(t, []string{"Carrefour", "Leclerc", "Monoprix"}, info.Stores)
}

func TestLookupByBarcode_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not json"))
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := looker.LookupByBarcode(context.Background(), "3333333333333")
	assert.Error(t, err)
}

func TestLookupByBarcode_URLConstruction(t *testing.T) {
	var capturedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		response := offResponse{
			Status: 1,
			Product: offProduct{
				Name:   "Test Product",
				Brands: "Brand",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer srv.Close()

	looker := &Looker{
		httpClient: srv.Client(),
		baseURL:    srv.URL,
	}

	_, err := looker.LookupByBarcode(context.Background(), "4444444444444")
	require.NoError(t, err)
	assert.Equal(t, "/api/v2/product/4444444444444.json", capturedPath)
}

func TestNewLooker(t *testing.T) {
	looker := NewLooker()
	assert.NotNil(t, looker)
	assert.NotNil(t, looker.httpClient)
	assert.Equal(t, offBaseURL, looker.baseURL)
}
