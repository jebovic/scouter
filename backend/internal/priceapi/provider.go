package priceapi

import (
	"context"
	"time"
)

// PriceListing is a normalized price result from any provider.
type PriceListing struct {
	Source      string    `json:"source"`
	ProductName string    `json:"productName"`
	Brand       string    `json:"brand"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	URL         string    `json:"url"`
	ImageURL    string    `json:"imageUrl"`
	Condition   string    `json:"condition"`
	Stores      string    `json:"stores"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProviderQuery is the input for provider searches.
type ProviderQuery struct {
	ProductName string
	Category    string
	Locale      string
	MaxResults  int
}

// PriceProvider is the interface each price source adapter implements.
type PriceProvider interface {
	Name() string
	Supports(category string) bool
	Search(ctx context.Context, q ProviderQuery) ([]PriceListing, error)
}

// AggregatedResult is the response from the Aggregator.
type AggregatedResult struct {
	Listings  []PriceListing `json:"listings"`
	Providers []string       `json:"providers"`
	Cached    bool           `json:"cached"`
}
