// Package negotiation implements the NegotiationAgent, which uses LLM tool-use to
// generate a structured negotiation script for a purchase option. It owns prompt
// construction, tool schema definition, and response parsing.
// The llm.Provider is transport-only.
package negotiation

import "fmt"

// NegotiationInput contains all context needed to generate a negotiation script.
type NegotiationInput struct {
	MissionName   string
	Category      string
	Budget        float64
	Currency      string
	OptionName    string
	OptionPrice   float64
	Attributes    map[string]string // key→value from option
	Constraints   []ConstraintInfo
	MarketPrices  []MarketPrice
	HistoricalLow *float64
}

// ConstraintInfo is a mission constraint relevant to the negotiation context.
type ConstraintInfo struct {
	Key   string
	Label string
	Value string
	Type  string // "hard" or "soft"
}

// MarketPrice is a price observation from a single merchant.
type MarketPrice struct {
	Merchant string
	Price    float64
}

// NegotiationScript is the structured output produced by the NegotiationAgent.
type NegotiationScript struct {
	OpeningOffer       float64  `json:"openingOffer"`
	WalkAwayPrice      float64  `json:"walkAwayPrice"`
	SuggestedDiscount  float64  `json:"suggestedDiscount"` // percentage 0–80
	TalkingPoints      []string `json:"talkingPoints"`
	CounterOfferScript []string `json:"counterOfferScript"` // step-by-step dialogue
	ConfidenceScore    float64  `json:"confidenceScore"`    // 0.0–1.0
}

// Validate checks that the NegotiationScript fields are logically consistent.
func (s NegotiationScript) Validate() error {
	if s.WalkAwayPrice < s.OpeningOffer {
		return fmt.Errorf("walkAwayPrice (%.2f) must be >= openingOffer (%.2f)", s.WalkAwayPrice, s.OpeningOffer)
	}
	if s.SuggestedDiscount < 0 || s.SuggestedDiscount > 80 {
		return fmt.Errorf("suggestedDiscount must be between 0 and 80, got %.2f", s.SuggestedDiscount)
	}
	if s.ConfidenceScore < 0 || s.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidenceScore must be between 0.0 and 1.0, got %.2f", s.ConfidenceScore)
	}
	if len(s.TalkingPoints) == 0 {
		return fmt.Errorf("talkingPoints must not be empty")
	}
	if len(s.CounterOfferScript) == 0 {
		return fmt.Errorf("counterOfferScript must not be empty")
	}
	return nil
}
