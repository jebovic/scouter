package settings

import (
	"encoding/json"
	"time"
)

// Setting is a single key-value configuration entry.
type Setting struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// UpdateRequest is a map of key -> raw JSON value for PATCH /api/settings.
type UpdateRequest map[string]json.RawMessage
