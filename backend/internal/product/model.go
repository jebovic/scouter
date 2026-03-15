package product

// ProductInfo is returned from barcode lookup.
type ProductInfo struct {
	Barcode    string   `json:"barcode"`
	Name       string   `json:"name"`
	Brand      string   `json:"brand"`
	Category   string   `json:"category"`
	ImageURL   string   `json:"imageUrl"`
	NutriScore string   `json:"nutriScore"` // A-E, only for food
	EcoScore   string   `json:"ecoScore"`   // A-E, only for food
	Stores     []string `json:"stores"`     // where it's sold
	Quantity   string   `json:"quantity"`   // e.g. "200g"
	Source     string   `json:"source"`     // "openfoodfacts"
}
