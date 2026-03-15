package priceapi

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- Mock provider ---

type mockProvider struct {
	name       string
	supports   bool
	listings   []PriceListing
	err        error
	callCount  int
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Supports(_ string) bool { return m.supports }
func (m *mockProvider) Search(_ context.Context, _ ProviderQuery) ([]PriceListing, error) {
	m.callCount++
	if m.err != nil {
		return nil, m.err
	}
	return m.listings, nil
}

// --- Tests ---

func TestAggregator_SingleProvider_CacheMiss(t *testing.T) {
	prov := &mockProvider{
		name:     "mock",
		supports: true,
		listings: []PriceListing{
			{Source: "mock", ProductName: "Widget", Price: 5.00},
		},
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{prov}, cache, 2)

	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Listings) != 1 {
		t.Fatalf("expected 1 listing, got %d", len(result.Listings))
	}
	if result.Cached {
		t.Error("expected Cached=false on first call")
	}
	if prov.callCount != 1 {
		t.Errorf("expected provider called once, got %d", prov.callCount)
	}

	// Verify cache was populated.
	key := cache.Key(prov.Name(), "Widget", "food")
	_, ok := cache.Get(key)
	if !ok {
		t.Error("expected cache to be populated after search")
	}
}

func TestAggregator_SingleProvider_CacheHit(t *testing.T) {
	prov := &mockProvider{
		name:     "mock",
		supports: true,
		listings: []PriceListing{
			{Source: "mock", ProductName: "Widget", Price: 5.00},
		},
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{prov}, cache, 2)

	// Prime cache.
	_, _ = agg.Search(context.Background(), "Widget", "food", "en")
	prov.callCount = 0 // reset

	// Second call should hit cache.
	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Cached {
		t.Error("expected Cached=true on second call")
	}
	if prov.callCount != 0 {
		t.Errorf("expected provider NOT called on cache hit, got %d calls", prov.callCount)
	}
}

func TestAggregator_MultipleProviders(t *testing.T) {
	prov1 := &mockProvider{
		name:     "prov1",
		supports: true,
		listings: []PriceListing{{Source: "prov1", ProductName: "Widget", Price: 3.00}},
	}
	prov2 := &mockProvider{
		name:     "prov2",
		supports: true,
		listings: []PriceListing{{Source: "prov2", ProductName: "Widget", Price: 4.00}},
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{prov1, prov2}, cache, 2)

	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Listings) != 2 {
		t.Fatalf("expected 2 listings, got %d", len(result.Listings))
	}
	if len(result.Providers) != 2 {
		t.Errorf("expected 2 providers in result, got %d", len(result.Providers))
	}
}

func TestAggregator_ProviderError_Isolated(t *testing.T) {
	good := &mockProvider{
		name:     "good",
		supports: true,
		listings: []PriceListing{{Source: "good", ProductName: "Widget", Price: 2.00}},
	}
	bad := &mockProvider{
		name:     "bad",
		supports: true,
		err:      errors.New("network failure"),
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{good, bad}, cache, 2)

	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Listings) != 1 {
		t.Fatalf("expected 1 listing from good provider, got %d", len(result.Listings))
	}
	if result.Listings[0].Source != "good" {
		t.Errorf("expected listing from 'good' provider, got %q", result.Listings[0].Source)
	}
}

func TestAggregator_AllProvidersFail(t *testing.T) {
	prov := &mockProvider{
		name:     "failing",
		supports: true,
		err:      errors.New("unavailable"),
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{prov}, cache, 2)

	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result even when all providers fail")
	}
	if len(result.Listings) != 0 {
		t.Errorf("expected empty listings, got %d", len(result.Listings))
	}
}

func TestAggregator_NoSupportingProviders(t *testing.T) {
	prov := &mockProvider{
		name:     "electronics",
		supports: false, // does not support the requested category
		listings: []PriceListing{{Source: "electronics", Price: 100.00}},
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{prov}, cache, 2)

	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Listings) != 0 {
		t.Errorf("expected empty listings when no providers support category, got %d", len(result.Listings))
	}
	if prov.callCount != 0 {
		t.Errorf("expected provider not called when Supports=false, got %d calls", prov.callCount)
	}
}

func TestAggregator_SortPriceZeroLast(t *testing.T) {
	prov := &mockProvider{
		name:     "mixed",
		supports: true,
		listings: []PriceListing{
			{Source: "mixed", ProductName: "C", Price: 0},
			{Source: "mixed", ProductName: "A", Price: 5.00},
			{Source: "mixed", ProductName: "B", Price: 2.50},
			{Source: "mixed", ProductName: "D", Price: 0},
		},
	}
	cache := NewCache(5 * time.Minute)
	agg := NewAggregator([]PriceProvider{prov}, cache, 2)

	result, err := agg.Search(context.Background(), "Widget", "food", "en")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Listings) != 4 {
		t.Fatalf("expected 4 listings, got %d", len(result.Listings))
	}

	// Price-positive entries first (ascending), then zeros.
	if result.Listings[0].Price != 2.50 {
		t.Errorf("listings[0] price: want 2.50, got %f", result.Listings[0].Price)
	}
	if result.Listings[1].Price != 5.00 {
		t.Errorf("listings[1] price: want 5.00, got %f", result.Listings[1].Price)
	}
	if result.Listings[2].Price != 0 {
		t.Errorf("listings[2] price: want 0, got %f", result.Listings[2].Price)
	}
	if result.Listings[3].Price != 0 {
		t.Errorf("listings[3] price: want 0, got %f", result.Listings[3].Price)
	}
}
