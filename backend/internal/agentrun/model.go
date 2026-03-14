// Package agentrun records every LLM agent invocation with input/result snapshots
// and computed diffs for the feedback loop history UI.
package agentrun

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AgentType identifies which agent produced the run.
type AgentType string

const (
	AgentTypeResearch AgentType = "research"
	AgentTypePricing  AgentType = "pricing"
)

// Run is a record of a single LLM agent invocation.
type Run struct {
	ID             uuid.UUID       `json:"id"`
	MissionID      uuid.UUID       `json:"missionId"`
	AgentType      AgentType       `json:"agentType"`
	InputSnapshot  json.RawMessage `json:"inputSnapshot"`
	ResultSnapshot json.RawMessage `json:"resultSnapshot"`
	Diff           json.RawMessage `json:"diff"`
	Feedback       string          `json:"feedback,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// DiffEntry holds the diff status of a single named item between two runs.
type DiffEntry struct {
	ID     string `json:"id,omitempty"`
	Name   string `json:"name"`
	Status string `json:"status"` // "new" | "removed" | "changed"
}

// Diff captures the computed difference between consecutive agent runs.
type Diff struct {
	Added   []DiffEntry `json:"added"`
	Removed []DiffEntry `json:"removed"`
	Changed []DiffEntry `json:"changed"`
}

// SnapshotItem is a minimal representation stored in result_snapshot.
type SnapshotItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
	Badge    string `json:"badge,omitempty"`
}

// ResultSnapshot is stored as result_snapshot JSON.
type ResultSnapshot struct {
	Items []SnapshotItem `json:"items"`
}

// InputSnapshotSummary is stored as input_snapshot JSON.
type InputSnapshotSummary struct {
	MissionName string         `json:"missionName"`
	Feedback    string         `json:"feedback,omitempty"`
	Pinned      []SnapshotItem `json:"pinned,omitempty"`
	Rejected    []SnapshotItem `json:"rejected,omitempty"`
}
