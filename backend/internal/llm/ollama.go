package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaProvider implements Provider using the Ollama native /api/chat endpoint.
// Supports tool use for models that have function calling capability (e.g. qwen2.5, llama3.1).
type OllamaProvider struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaProvider creates a new OllamaProvider.
func NewOllamaProvider(baseURL, model string) *OllamaProvider {
	return &OllamaProvider{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 180 * time.Second},
	}
}

// ── Ollama native /api/chat request/response types ───────────────────────────

type ollamaMessage struct {
	Role      string             `json:"role"`
	Content   string             `json:"content,omitempty"`
	ToolCalls []ollamaToolCall   `json:"tool_calls,omitempty"`
}

type ollamaToolCall struct {
	Function ollamaToolCallFunc `json:"function"`
}

type ollamaToolCallFunc struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ollamaTool struct {
	Type     string           `json:"type"` // always "function"
	Function ollamaToolSchema `json:"function"`
}

type ollamaToolSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type ollamaChatResponse struct {
	Message           ollamaMessage `json:"message"`
	PromptEvalCount   int           `json:"prompt_eval_count"`
	EvalCount         int           `json:"eval_count"`
}

// Complete sends a chat completion request to the Ollama /api/chat endpoint.
// Tool use is supported for compatible models.
func (p *OllamaProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	msgs := make([]ollamaMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		msgs = append(msgs, ollamaMessage{Role: m.Role, Content: m.Content})
	}

	tools := make([]ollamaTool, 0, len(req.Tools))
	for _, t := range req.Tools {
		// Ollama expects the full JSON Schema under "parameters" (OpenAI format).
		// InputSchema already contains "type", "properties", "required" — pass it through.
		params := map[string]any{
			"type": "object",
		}
		if props, ok := t.InputSchema["properties"]; ok {
			params["properties"] = props
		}
		if req, ok := t.InputSchema["required"]; ok {
			params["required"] = req
		}

		tools = append(tools, ollamaTool{
			Type: "function",
			Function: ollamaToolSchema{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  params,
			},
		})
	}

	body, err := json.Marshal(ollamaChatRequest{
		Model:    p.model,
		Messages: msgs,
		Tools:    tools,
		Stream:   false,
	})
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("create ollama request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("ollama request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return CompletionResponse{}, &StatusError{Code: httpResp.StatusCode}
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&ollamaResp); err != nil {
		return CompletionResponse{}, fmt.Errorf("decode ollama response: %w", err)
	}

	resp := CompletionResponse{
		Content: ollamaResp.Message.Content,
		Usage: Usage{
			InputTokens:  ollamaResp.PromptEvalCount,
			OutputTokens: ollamaResp.EvalCount,
			Provider:     "ollama",
			Model:        p.model,
		},
	}

	for _, tc := range ollamaResp.Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, ToolCall{
			Name:  tc.Function.Name,
			Input: tc.Function.Arguments,
		})
	}

	return resp, nil
}
