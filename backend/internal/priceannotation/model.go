package priceannotation

import "time"

// AnnotationType classifies a price timeline event.
type AnnotationType string

const (
	TypeNewLow   AnnotationType = "new_low"
	TypeRebound  AnnotationType = "rebound"
	TypeSpikeUp  AnnotationType = "spike_up"
	TypeSpikeDown AnnotationType = "spike_down"
	TypeStable   AnnotationType = "stable"
)

// Annotation describes a notable event in a price history series.
type Annotation struct {
	RecordedAt  time.Time      `json:"recordedAt"`
	Price       float64        `json:"price"`
	Type        AnnotationType `json:"type"`
	Label       string         `json:"label"`
	Description string         `json:"description"`
}

// PriceAnnotations is the response payload for the price annotations endpoint.
type PriceAnnotations struct {
	ItemID      string       `json:"itemId"`
	Annotations []Annotation `json:"annotations"`
}
