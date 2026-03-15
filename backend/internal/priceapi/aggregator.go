package priceapi

import (
	"context"
	"log"
	"sort"
	"sync"
)

// Aggregator fans out search requests to multiple PriceProviders in parallel,
// merges and deduplicates the results, and caches per-provider responses.
type Aggregator struct {
	providers []PriceProvider
	cache     *Cache
	sem       chan struct{}
}

// NewAggregator creates a new Aggregator with the given providers, cache, and
// concurrency limit.
func NewAggregator(providers []PriceProvider, cache *Cache, maxConcurrent int) *Aggregator {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	return &Aggregator{
		providers: providers,
		cache:     cache,
		sem:       make(chan struct{}, maxConcurrent),
	}
}

// Search queries all supporting providers in parallel and returns a merged,
// sorted AggregatedResult. Provider errors are logged and do not propagate.
func (a *Aggregator) Search(ctx context.Context, productName, category, locale string) (*AggregatedResult, error) {
	// Filter providers that support the requested category.
	var supporting []PriceProvider
	for _, p := range a.providers {
		if p.Supports(category) {
			supporting = append(supporting, p)
		}
	}

	if len(supporting) == 0 {
		return &AggregatedResult{
			Listings:  []PriceListing{},
			Providers: []string{},
		}, nil
	}

	type providerResult struct {
		name     string
		listings []PriceListing
		cached   bool
	}

	results := make(chan providerResult, len(supporting))
	var wg sync.WaitGroup

	for _, prov := range supporting {
		wg.Add(1)
		go func(p PriceProvider) {
			defer wg.Done()

			key := a.cache.Key(p.Name(), productName, category)

			// Check cache first (no semaphore needed — in-memory read).
			if cached, ok := a.cache.Get(key); ok {
				results <- providerResult{name: p.Name(), listings: cached, cached: true}
				return
			}

			// Acquire concurrency slot (respect context cancellation to avoid goroutine leak).
			select {
			case a.sem <- struct{}{}:
			case <-ctx.Done():
				results <- providerResult{name: p.Name(), listings: nil}
				return
			}
			defer func() { <-a.sem }()

			listings, err := p.Search(ctx, ProviderQuery{
				ProductName: productName,
				Category:    category,
				Locale:      locale,
				MaxResults:  10,
			})
			if err != nil {
				log.Printf("priceapi: provider %q error: %v", p.Name(), err)
				results <- providerResult{name: p.Name(), listings: nil}
				return
			}

			a.cache.Set(key, listings)
			results <- providerResult{name: p.Name(), listings: listings, cached: false}
		}(prov)
	}

	wg.Wait()
	close(results)

	var allListings []PriceListing
	providerNames := make([]string, 0, len(supporting))
	allCached := true

	for res := range results {
		providerNames = append(providerNames, res.name)
		if !res.cached {
			allCached = false
		}
		allListings = append(allListings, res.listings...)
	}

	// Sort: price > 0 ascending, then price = 0 at end.
	sort.SliceStable(allListings, func(i, j int) bool {
		pi, pj := allListings[i].Price, allListings[j].Price
		switch {
		case pi > 0 && pj > 0:
			return pi < pj
		case pi > 0:
			return true
		case pj > 0:
			return false
		default:
			return false
		}
	})

	if allListings == nil {
		allListings = []PriceListing{}
	}

	return &AggregatedResult{
		Listings:  allListings,
		Providers: providerNames,
		Cached:    allCached && len(supporting) > 0,
	}, nil
}
