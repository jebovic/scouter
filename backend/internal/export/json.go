package export

import (
	"encoding/json"
	"fmt"
)

// jsonExport is the versioned JSON envelope.
type jsonExport struct {
	Version string `json:"version"`
	*ExportData
}

// ToJSON serializes ExportData to indented JSON with a version envelope.
func ToJSON(data *ExportData) ([]byte, error) {
	envelope := jsonExport{
		Version:    "1.0",
		ExportData: data,
	}
	b, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("export json: %w", err)
	}
	return b, nil
}
