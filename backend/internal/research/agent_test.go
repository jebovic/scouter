package research

import (
	"testing"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/option"
)

func TestParseToolResponse_NoToolCall(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "wrong_tool", Input: map[string]any{}},
		},
	}
	_, err := parseToolResponse(resp)
	if err == nil {
		t.Error("parseToolResponse with wrong tool name should return error")
	}
}

func TestParseToolResponse_EmptyToolCalls(t *testing.T) {
	resp := llm.CompletionResponse{ToolCalls: nil}
	_, err := parseToolResponse(resp)
	if err == nil {
		t.Error("parseToolResponse with no tool calls should return error")
	}
}

func TestUnmarshalOptions(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]any
		wantLen int
		wantErr bool
	}{
		{
			name: "two valid options",
			input: map[string]any{
				"options": []any{
					map[string]any{
						"name":     "MacBook Pro M3",
						"category": "laptop",
						"badge":    "recommended",
						"notes":    "Best for creative pros",
						"warnings": []any{"High price"},
					},
					map[string]any{
						"name":     "Dell XPS 15",
						"category": "laptop",
						"badge":    "alternative",
					},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "empty options list",
			input:   map[string]any{"options": []any{}},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "invalid structure",
			input:   map[string]any{"options": "not-an-array"},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, err := unmarshalOptions(tt.input)
			if tt.wantErr && err == nil {
				t.Error("unmarshalOptions() should return error")
				return
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unmarshalOptions() error: %v", err)
				return
			}
			if !tt.wantErr && len(opts) != tt.wantLen {
				t.Errorf("len(opts) = %d; want %d", len(opts), tt.wantLen)
			}
		})
	}
}

func TestUnmarshalOptions_NilSlicesBecomEmpty(t *testing.T) {
	input := map[string]any{
		"options": []any{
			map[string]any{
				"name":     "ThinkPad X1",
				"category": "laptop",
				"badge":    "watch",
				// no attributes or warnings — should default to empty slice
			},
		},
	}

	opts, err := unmarshalOptions(input)
	if err != nil {
		t.Fatalf("unmarshalOptions() error: %v", err)
	}
	if len(opts) != 1 {
		t.Fatalf("expected 1 option, got %d", len(opts))
	}
	if opts[0].Attributes == nil {
		t.Error("Attributes = nil; want empty slice")
	}
	if opts[0].Warnings == nil {
		t.Error("Warnings = nil; want empty slice")
	}
}

func TestUnmarshalOptions_WithPriceRange(t *testing.T) {
	input := map[string]any{
		"options": []any{
			map[string]any{
				"name":     "Sony WH-1000XM5",
				"category": "audio",
				"badge":    "recommended",
				"price_range": map[string]any{
					"min":  float64(280),
					"max":  float64(380),
					"best": float64(299),
				},
			},
		},
	}

	opts, err := unmarshalOptions(input)
	if err != nil {
		t.Fatalf("unmarshalOptions() error: %v", err)
	}
	if opts[0].PriceRange == nil {
		t.Fatal("PriceRange = nil; want non-nil")
	}
	if opts[0].PriceRange.Best != 299 {
		t.Errorf("PriceRange.Best = %.0f; want 299", opts[0].PriceRange.Best)
	}
}

func TestParseToolResponse_ValidCall(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{
				Name: "submit_research_options",
				Input: map[string]any{
					"options": []any{
						map[string]any{
							"name":     "Test Option",
							"category": "test",
							"badge":    "watch",
						},
					},
				},
			},
		},
	}

	opts, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse() error: %v", err)
	}
	if len(opts) != 1 {
		t.Errorf("len(opts) = %d; want 1", len(opts))
	}
	if opts[0].Badge != "watch" {
		t.Errorf("Badge = %q; want %q", opts[0].Badge, "watch")
	}
}

func TestParseToolResponse_SkipsUnrelatedToolCalls(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "some_other_tool", Input: map[string]any{}},
			{
				Name: "submit_research_options",
				Input: map[string]any{
					"options": []any{
						map[string]any{"name": "X", "category": "c", "badge": "watch"},
					},
				},
			},
		},
	}

	opts, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse() error: %v", err)
	}
	if len(opts) != 1 {
		t.Errorf("len(opts) = %d; want 1", len(opts))
	}
}

// Verify option.Attribute fields are correctly unmarshaled.
func TestUnmarshalOptions_Attributes(t *testing.T) {
	input := map[string]any{
		"options": []any{
			map[string]any{
				"name":     "Laptop",
				"category": "computing",
				"badge":    "alternative",
				"attributes": []any{
					map[string]any{
						"key":   "ram",
						"label": "RAM",
						"value": "16GB",
						"type":  "text",
					},
				},
			},
		},
	}

	opts, err := unmarshalOptions(input)
	if err != nil {
		t.Fatalf("unmarshalOptions() error: %v", err)
	}
	if len(opts[0].Attributes) != 1 {
		t.Fatalf("Attributes len = %d; want 1", len(opts[0].Attributes))
	}
	attr := opts[0].Attributes[0]
	if attr.Key != "ram" {
		t.Errorf("Attribute.Key = %q; want %q", attr.Key, "ram")
	}
	if attr.Type != "text" {
		t.Errorf("Attribute.Type = %q; want %q", attr.Type, "text")
	}
}

// Compile-time check: option.Option fields used in unmarshalOptions must exist.
var _ option.Option
