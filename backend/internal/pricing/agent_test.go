package pricing

import (
	"testing"

	"github.com/jibei/scouter/internal/llm"
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

func TestUnmarshalItems_TwoValidItems(t *testing.T) {
	input := map[string]any{
		"items": []any{
			map[string]any{
				"name":          "Sony WH-1000XM5",
				"merchant":      "Amazon",
				"cost_category": "audio",
				"price":         float64(299),
				"status":        "buy",
				"note":          "Best deal right now",
				"url":           "https://example.com/sony",
			},
			map[string]any{
				"name":          "Apple AirPods Pro",
				"merchant":      "Fnac",
				"cost_category": "audio",
				"price":         float64(249),
				"status":        "watch",
			},
		},
	}

	items, err := unmarshalItems(input)
	if err != nil {
		t.Fatalf("unmarshalItems() error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d; want 2", len(items))
	}
	if items[0].Name != "Sony WH-1000XM5" {
		t.Errorf("items[0].Name = %q; want %q", items[0].Name, "Sony WH-1000XM5")
	}
	if items[0].Merchant != "Amazon" {
		t.Errorf("items[0].Merchant = %q; want %q", items[0].Merchant, "Amazon")
	}
	if items[0].Price != 299 {
		t.Errorf("items[0].Price = %.0f; want 299", items[0].Price)
	}
	if items[0].Status != "buy" {
		t.Errorf("items[0].Status = %q; want %q", items[0].Status, "buy")
	}
}

func TestUnmarshalItems_EmptyList(t *testing.T) {
	input := map[string]any{"items": []any{}}
	items, err := unmarshalItems(input)
	if err != nil {
		t.Fatalf("unmarshalItems() error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("len(items) = %d; want 0", len(items))
	}
}

func TestUnmarshalItems_WithOriginalEstimate(t *testing.T) {
	estimate := float64(350)
	input := map[string]any{
		"items": []any{
			map[string]any{
				"name":              "MacBook Pro",
				"merchant":          "Apple Store",
				"cost_category":     "computing",
				"price":             float64(1299),
				"original_estimate": estimate,
				"status":            "crisis",
			},
		},
	}

	items, err := unmarshalItems(input)
	if err != nil {
		t.Fatalf("unmarshalItems() error: %v", err)
	}
	if items[0].OriginalEstimate == nil {
		t.Fatal("OriginalEstimate = nil; want non-nil")
	}
	if *items[0].OriginalEstimate != 350 {
		t.Errorf("OriginalEstimate = %.0f; want 350", *items[0].OriginalEstimate)
	}
}

func TestUnmarshalItems_NilOriginalEstimate(t *testing.T) {
	input := map[string]any{
		"items": []any{
			map[string]any{
				"name":          "GPU",
				"merchant":      "Cdiscount",
				"cost_category": "hardware",
				"price":         float64(799),
				"status":        "watch",
			},
		},
	}

	items, err := unmarshalItems(input)
	if err != nil {
		t.Fatalf("unmarshalItems() error: %v", err)
	}
	if items[0].OriginalEstimate != nil {
		t.Errorf("OriginalEstimate = %v; want nil", *items[0].OriginalEstimate)
	}
}

func TestParseToolResponse_ValidCall(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{
				Name: "submit_pricing_items",
				Input: map[string]any{
					"items": []any{
						map[string]any{
							"name":          "Test Item",
							"merchant":      "Test Shop",
							"cost_category": "test",
							"price":         float64(100),
							"status":        "watch",
						},
					},
				},
			},
		},
	}

	items, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse() error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d; want 1", len(items))
	}
	if items[0].Status != "watch" {
		t.Errorf("Status = %q; want %q", items[0].Status, "watch")
	}
}

func TestParseToolResponse_SkipsUnrelatedToolCalls(t *testing.T) {
	resp := llm.CompletionResponse{
		ToolCalls: []llm.ToolCall{
			{Name: "some_other_tool", Input: map[string]any{}},
			{
				Name: "submit_pricing_items",
				Input: map[string]any{
					"items": []any{
						map[string]any{
							"name":          "X",
							"merchant":      "Y",
							"cost_category": "z",
							"price":         float64(1),
							"status":        "buy",
						},
					},
				},
			},
		},
	}

	items, err := parseToolResponse(resp)
	if err != nil {
		t.Fatalf("parseToolResponse() error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d; want 1", len(items))
	}
}

func TestUnmarshalItems_AllFieldsPreserved(t *testing.T) {
	input := map[string]any{
		"items": []any{
			map[string]any{
				"name":          "RTX 4080",
				"merchant":      "LDLC",
				"cost_category": "hardware",
				"price":         float64(1199),
				"status":        "flash-sale",
				"note":          "Limited stock, -20% this weekend",
				"url":           "https://ldlc.com/rtx4080",
			},
		},
	}

	items, err := unmarshalItems(input)
	if err != nil {
		t.Fatalf("unmarshalItems() error: %v", err)
	}
	item := items[0]
	if item.CostCategory != "hardware" {
		t.Errorf("CostCategory = %q; want %q", item.CostCategory, "hardware")
	}
	if item.Note != "Limited stock, -20% this weekend" {
		t.Errorf("Note = %q; want different", item.Note)
	}
	if item.URL != "https://ldlc.com/rtx4080" {
		t.Errorf("URL = %q; want different", item.URL)
	}
}
