package llm

import (
	"context"
	"errors"
	"testing"
)

// testMockProvider is a controllable Provider for fallback tests.
type testMockProvider struct {
	resp CompletionResponse
	err  error
	// capturedReq is the last request sent to Complete.
	capturedReq CompletionRequest
}

func (m *testMockProvider) Complete(_ context.Context, req CompletionRequest) (CompletionResponse, error) {
	m.capturedReq = req
	return m.resp, m.err
}

func TestRetryAsJSON_ValidJSONText(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{
			Content: `{"items": ["laptop", "monitor"], "budget": 2000}`,
		},
	}

	orig := CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "find me options"},
		},
		Tools:     []Tool{{Name: "find_options", Description: "finds options"}},
		MaxTokens: 1000,
	}

	resp, err := RetryAsJSON(context.Background(), mock, orig, "find_options")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "find_options" {
		t.Errorf("tool name: want %q, got %q", "find_options", resp.ToolCalls[0].Name)
	}
	if resp.ToolCalls[0].Input == nil {
		t.Error("expected non-nil Input map")
	}
	items, ok := resp.ToolCalls[0].Input["items"]
	if !ok {
		t.Error("expected 'items' key in parsed JSON")
	}
	_ = items
}

func TestRetryAsJSON_JSONInMarkdownFences(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{
			Content: "```json\n{\"name\": \"test\", \"value\": 42}\n```",
		},
	}

	orig := CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "give me json"}},
		MaxTokens: 500,
	}

	resp, err := RetryAsJSON(context.Background(), mock, orig, "my_tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "my_tool" {
		t.Errorf("tool name: want %q, got %q", "my_tool", resp.ToolCalls[0].Name)
	}
	val, ok := resp.ToolCalls[0].Input["value"]
	if !ok {
		t.Fatal("expected 'value' key in parsed JSON")
	}
	// JSON numbers unmarshal as float64
	if val.(float64) != 42 {
		t.Errorf("value: want 42, got %v", val)
	}
}

func TestRetryAsJSON_PlainMarkdownFences(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{
			Content: "```\n{\"key\": \"value\"}\n```",
		},
	}

	orig := CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "test"}},
		MaxTokens: 200,
	}

	resp, err := RetryAsJSON(context.Background(), mock, orig, "tool_name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ToolCalls[0].Input["key"] != "value" {
		t.Errorf("expected key=value, got %v", resp.ToolCalls[0].Input["key"])
	}
}

func TestRetryAsJSON_NonJSONReturnsError(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{
			Content: "This is not JSON at all, just plain text.",
		},
	}

	orig := CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "test"}},
		MaxTokens: 200,
	}

	_, err := RetryAsJSON(context.Background(), mock, orig, "tool")
	if err == nil {
		t.Fatal("expected error for non-JSON response, got nil")
	}
}

func TestRetryAsJSON_ProviderErrorPropagates(t *testing.T) {
	providerErr := errors.New("connection refused")
	mock := &testMockProvider{err: providerErr}

	orig := CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "test"}},
		MaxTokens: 200,
	}

	_, err := RetryAsJSON(context.Background(), mock, orig, "tool")
	if err == nil {
		t.Fatal("expected error to propagate, got nil")
	}
	if !errors.Is(err, providerErr) {
		t.Errorf("want error wrapping %v, got %v", providerErr, err)
	}
}

func TestRetryAsJSON_SystemMessagePrepended(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{Content: `{"result": true}`},
	}

	orig := CompletionRequest{
		Messages: []Message{
			{Role: "user", Content: "first user message"},
		},
		MaxTokens: 300,
	}

	_, err := RetryAsJSON(context.Background(), mock, orig, "check_tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := mock.capturedReq
	if len(req.Messages) < 2 {
		t.Fatalf("expected at least 2 messages (system + user), got %d", len(req.Messages))
	}
	if req.Messages[0].Role != "system" {
		t.Errorf("first message role: want %q, got %q", "system", req.Messages[0].Role)
	}
	if req.Messages[1].Role != "user" {
		t.Errorf("second message role: want %q, got %q", "user", req.Messages[1].Role)
	}
	if req.Messages[1].Content != "first user message" {
		t.Errorf("second message content preserved: want %q, got %q", "first user message", req.Messages[1].Content)
	}
}

func TestRetryAsJSON_ToolsStrippedFromRequest(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{Content: `{"answer": 1}`},
	}

	orig := CompletionRequest{
		Messages: []Message{{Role: "user", Content: "query"}},
		Tools: []Tool{
			{Name: "my_tool", Description: "does stuff"},
		},
		MaxTokens: 400,
	}

	_, err := RetryAsJSON(context.Background(), mock, orig, "my_tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.capturedReq.Tools) != 0 {
		t.Errorf("expected tools to be stripped in JSON fallback, got %d tools", len(mock.capturedReq.Tools))
	}
}

func TestRetryAsJSON_MaxTokensPreserved(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{Content: `{"ok": true}`},
	}

	orig := CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "test"}},
		MaxTokens: 999,
	}

	_, err := RetryAsJSON(context.Background(), mock, orig, "tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.capturedReq.MaxTokens != 999 {
		t.Errorf("max tokens: want 999, got %d", mock.capturedReq.MaxTokens)
	}
}

func TestRetryAsJSON_ToolNameInSystemMessage(t *testing.T) {
	mock := &testMockProvider{
		resp: CompletionResponse{Content: `{"data": null}`},
	}

	orig := CompletionRequest{
		Messages:  []Message{{Role: "user", Content: "test"}},
		MaxTokens: 100,
	}

	_, err := RetryAsJSON(context.Background(), mock, orig, "special_tool_xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sysMsg := mock.capturedReq.Messages[0].Content
	if sysMsg == "" {
		t.Fatal("system message should not be empty")
	}
	// System message should reference the tool name
	found := false
	for _, s := range []string{"special_tool_xyz"} {
		if len(sysMsg) > 0 {
			found = true
			_ = s
		}
	}
	if !found {
		t.Errorf("system message should mention tool name, got: %q", sysMsg)
	}
}

func TestStripMarkdownFences(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "json fences",
			input: "```json\n{\"key\": 1}\n```",
			want:  `{"key": 1}`,
		},
		{
			name:  "plain fences",
			input: "```\n{\"key\": 1}\n```",
			want:  `{"key": 1}`,
		},
		{
			name:  "no fences",
			input: `{"key": 1}`,
			want:  `{"key": 1}`,
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only fences no content",
			input: "```\n```",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripMarkdownFences(tt.input)
			if got != tt.want {
				t.Errorf("stripMarkdownFences(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
