package negotiationsim

import "time"

// NegotiationScript holds a complete negotiation script for a shopping item.
type NegotiationScript struct {
	ItemID           string       `json:"itemId"`
	ItemName         string       `json:"itemName"`
	CurrentPrice     float64      `json:"currentPrice"`
	TargetPrice      float64      `json:"targetPrice"`
	PotentialSavings float64      `json:"potentialSavings"`
	Scripts          []ScriptStep `json:"scripts"`
	Tips             []string     `json:"tips"`
	GeneratedAt      time.Time    `json:"generatedAt"`
}

// ScriptStep is a single dialogue turn in the negotiation script.
type ScriptStep struct {
	Phase   string `json:"phase"`   // "opening", "negotiation", "closing"
	Speaker string `json:"speaker"` // "vous", "vendeur"
	Text    string `json:"text"`    // French dialogue
	Tip     string `json:"tip"`     // coaching tip
}
