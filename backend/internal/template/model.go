package template

import "github.com/jibei/scouter/internal/mission"

// Template is a pre-built mission blueprint compiled into the binary.
// Templates are hardcoded — no database table, no LLM calls.
type Template struct {
	Slug            string                `json:"slug"`
	Name            string                `json:"name"`
	Icon            string                `json:"icon"`
	Category        string                `json:"category"`
	Description     string                `json:"description"`
	Constraints     []mission.Constraint  `json:"constraints"`
	CostCategories  []string              `json:"costCategories"`
	WeightProfile   mission.WeightProfile `json:"weightProfile"`
	SuggestedBudget *BudgetRange          `json:"suggestedBudget,omitempty"`
}

// BudgetRange provides a suggested min/max budget hint for a template.
// It is advisory only — the user sets the actual budget in MissionForm.
type BudgetRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}
