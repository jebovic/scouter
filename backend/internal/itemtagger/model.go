package itemtagger

import "time"

// ItemTags holds the auto-detected metadata for a shopping item.
type ItemTags struct {
	ItemID               string    `json:"itemId"`
	ItemName             string    `json:"itemName"`
	SuggestedCategory    string    `json:"suggestedCategory"`
	Tags                 []string  `json:"tags"`
	Brand                string    `json:"brand"`
	ProductType          string    `json:"productType"`
	IsLuxury             bool      `json:"isLuxury"`
	IsTech               bool      `json:"isTech"`
	WarrantyHint         string    `json:"warrantyHint"`
	RefurbishedAvailable bool      `json:"refurbishedAvailable"`
	GeneratedAt          time.Time `json:"generatedAt"`
}
