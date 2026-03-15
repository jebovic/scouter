// Package giftfinder provides hardcoded French gift catalog suggestions
// filtered by occasion, recipient, and budget. No LLM or external API calls
// required — all data is compiled into the binary.
package giftfinder

import "time"

// GiftSuggestion is a single gift idea returned to the client.
type GiftSuggestion struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	PriceRange     string   `json:"priceRange"`
	EstimatedPrice float64  `json:"estimatedPrice"`
	Description    string   `json:"description"`
	Occasion       string   `json:"occasion"`
	Recipient      string   `json:"recipient"`
	Platforms      []string `json:"platforms"`
	Score          int      `json:"score"`
}

// GiftResult is the top-level response envelope.
type GiftResult struct {
	MissionID   string           `json:"missionId"`
	Budget      float64          `json:"budget"`
	Suggestions []GiftSuggestion `json:"suggestions"`
	TotalFound  int              `json:"totalFound"`
	GeneratedAt time.Time        `json:"generatedAt"`
}

// catalogItem is the internal representation of a catalog entry.
// It may match multiple occasions and recipients.
type catalogItem struct {
	id             string
	name           string
	category       string
	priceRange     string
	estimatedPrice float64
	description    string
	occasions      []string
	recipients     []string
	platforms      []string
	keywords       []string // used for relevance scoring against mission name
}
