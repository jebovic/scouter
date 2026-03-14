package mission

import (
	"time"

	"github.com/google/uuid"
)

// WeightProfile configures how the ScoreEngine weights dimensions when scoring options.
// Values are relative weights (any positive number); they are normalized internally.
// Defaults to equal weight across all dimensions when empty.
type WeightProfile struct {
	Price   float64 `json:"price"`
	Quality float64 `json:"quality"`
	Feature float64 `json:"feature"`
}

// Mission represents a single spending research project.
type Mission struct {
	ID             uuid.UUID       `json:"id"`
	Slug           string          `json:"slug"`
	Name           string          `json:"name"`
	Icon           string          `json:"icon"`
	Category       string          `json:"category"` // travel | renovation | electronics | computing | custom
	Budget         float64         `json:"budget"`
	Currency       string          `json:"currency"`
	Locale         string          `json:"locale"`
	Phase          string          `json:"phase"` // researching | comparing | buying | done
	Constraints    []Constraint    `json:"constraints"`
	CostCategories []string        `json:"costCategories"`
	Timeline       []TimelineEvent `json:"timeline"`
	WeightProfile  WeightProfile   `json:"weightProfile"`
	ShareToken     *string         `json:"shareToken,omitempty"`
	ArchivedAt     *time.Time      `json:"archivedAt,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// Constraint is a hard or soft requirement that filters options.
type Constraint struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value any    `json:"value"`
	Type  string `json:"type"` // hard | soft
}

// TimelineEvent is a dated milestone or deadline.
type TimelineEvent struct {
	Date    string `json:"date"`
	Label   string `json:"label"`
	Urgent  bool   `json:"urgent"`
}

// CreateRequest is the payload for creating a new mission.
type CreateRequest struct {
	Name           string          `json:"name"           validate:"required,min=1,max=200"`
	Icon           string          `json:"icon"`
	Category       string          `json:"category"       validate:"required,oneof=travel renovation electronics computing custom"`
	Budget         float64         `json:"budget"         validate:"required,gt=0"`
	Currency       string          `json:"currency"`
	Locale         string          `json:"locale"`
	Constraints    []Constraint    `json:"constraints"`
	CostCategories []string        `json:"costCategories"`
}

// UpdateRequest is the payload for updating an existing mission.
type UpdateRequest struct {
	Name           *string          `json:"name,omitempty"          validate:"omitempty,min=1,max=200"`
	Icon           *string          `json:"icon,omitempty"`
	Budget         *float64         `json:"budget,omitempty"        validate:"omitempty,gt=0"`
	Phase          *string          `json:"phase,omitempty"         validate:"omitempty,oneof=researching comparing buying done"`
	Constraints    []Constraint     `json:"constraints,omitempty"`
	CostCategories []string         `json:"costCategories,omitempty"`
	Timeline       []TimelineEvent  `json:"timeline,omitempty"`
	WeightProfile  *WeightProfile   `json:"weightProfile,omitempty"`
}
