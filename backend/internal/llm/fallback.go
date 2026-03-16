package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// RetryAsJSON retries a failed tool-use request in JSON-only mode.
// It strips tools from the request, prepends a system instruction asking for
// raw JSON matching the named tool's schema, calls provider.Complete, then wraps
// the text response as a synthetic ToolCall with Name=toolName and parsed JSON input.
//
// The original request's messages are preserved verbatim. Callers (agents) own the
// original tool schema; this function only handles the retry envelope.
func RetryAsJSON(ctx context.Context, provider Provider, orig CompletionRequest, toolName string) (CompletionResponse, error) {
	// Build JSON-mode prompt
	sysMsg := fmt.Sprintf(
		"You MUST respond with a single raw JSON object only. "+
			"Do NOT use tool calls. Do NOT wrap in markdown. Do NOT add any explanation. "+
			"Output ONLY the JSON that matches the schema for %q.", toolName)

	msgs := make([]Message, 0, len(orig.Messages)+1)
	msgs = append(msgs, Message{Role: "system", Content: sysMsg})
	msgs = append(msgs, orig.Messages...)

	jsonReq := CompletionRequest{
		Messages:  msgs,
		MaxTokens: orig.MaxTokens,
		// Tools intentionally omitted
	}

	resp, err := provider.Complete(ctx, jsonReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("json fallback llm call: %w", err)
	}

	// If content is empty but the model still returned a tool call (e.g. deepseek
	// ignoring the JSON-only instruction), use the matching tool call directly.
	if strings.TrimSpace(resp.Content) == "" {
		for _, tc := range resp.ToolCalls {
			if tc.Name == toolName {
				return resp, nil
			}
		}
	}

	// Strip markdown fences if present
	raw := strings.TrimSpace(resp.Content)
	raw = stripMarkdownFences(raw)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return CompletionResponse{}, fmt.Errorf("json fallback: model returned non-JSON: %w", err)
	}

	resp.ToolCalls = []ToolCall{{Name: toolName, Input: parsed}}
	return resp, nil
}

// stripMarkdownFences removes leading/trailing markdown code fences
// (```json ... ``` or ``` ... ```). Returns trimmed inner content.
func stripMarkdownFences(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}

	// Remove the opening fence line (```json or ```).
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[idx+1:]
	} else {
		// Opening fence with no newline — nothing left.
		return ""
	}

	// Remove the closing ``` (last line).
	if strings.HasSuffix(s, "```") {
		if idx := strings.LastIndex(s, "\n"); idx != -1 {
			s = s[:idx]
		} else {
			// The entire remaining content is just "```".
			return ""
		}
	}

	return strings.TrimSpace(s)
}
