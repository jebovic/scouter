// Package autotag implements Smart Category Auto-Tagging using the LLM.
// It suggests a French category label for a mission based on its name,
// description, and constraints. A deterministic keyword-matching fallback
// is used when the LLM is unavailable.
package autotag

// CategorySuggestion is the response returned by the suggest-category endpoint.
type CategorySuggestion struct {
	Category     string   `json:"category"`
	Confidence   float64  `json:"confidence"` // 0–1
	Alternatives []string `json:"alternatives"`
}

// TagInput carries the context needed to determine a category.
type TagInput struct {
	MissionSlug string
	Name        string
	Constraints []string // constraint labels, may be empty
}

// Categories is the canonical French taxonomy used for tagging.
var Categories = []string{
	"Électronique", "Informatique", "Audiovisuel",
	"Alimentation", "Cuisine", "Boissons",
	"Voyage", "Transport", "Hébergement",
	"Mode", "Vêtements", "Chaussures",
	"Maison", "Mobilier", "Décoration",
	"Sport", "Loisirs", "Jeux",
	"Santé", "Beauté", "Bien-être",
	"Auto", "Moto", "Vélo",
	"Livres", "Culture", "Musique",
	"Services", "Abonnements", "Autre",
}
