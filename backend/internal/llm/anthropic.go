package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

// AnthropicProvider implements Provider using the Anthropic SDK.
// It uses tool use (function calling) to produce structured output.
type AnthropicProvider struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewAnthropicProvider creates a new AnthropicProvider with the given API key.
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
		model:  anthropic.ModelClaudeSonnet4_6,
	}
}

// Complete sends a completion request to the Anthropic API.
func (p *AnthropicProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	// Convert messages
	messages := make([]anthropic.MessageParam, 0, len(req.Messages))
	for _, m := range req.Messages {
		switch m.Role {
		case "user":
			messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		case "assistant":
			messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}

	// Convert tools
	tools := make([]anthropic.ToolUnionParam, 0, len(req.Tools))
	for _, t := range req.Tools {
		schema := anthropic.ToolInputSchemaParam{
			Properties: t.InputSchema["properties"],
		}
		if required, ok := t.InputSchema["required"].([]string); ok {
			schema.Required = required
		}

		toolParam := anthropic.ToolParam{
			Name:        t.Name,
			Description: param.NewOpt(t.Description),
			InputSchema: schema,
		}
		tools = append(tools, anthropic.ToolUnionParam{OfTool: &toolParam})
	}

	maxTokens := int64(req.MaxTokens)
	if maxTokens == 0 {
		maxTokens = 4096
	}

	sdkParams := anthropic.MessageNewParams{
		Model:     p.model,
		Messages:  messages,
		MaxTokens: maxTokens,
	}
	if len(tools) > 0 {
		sdkParams.Tools = tools
	}

	msg, err := p.client.Messages.New(ctx, sdkParams)
	if err != nil {
		return CompletionResponse{}, fmt.Errorf("anthropic completion: %w", err)
	}

	resp := CompletionResponse{
		Usage: Usage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
			Provider:     "anthropic",
			Model:        string(p.model),
		},
	}
	var sb strings.Builder
	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			sb.WriteString(block.Text)
		case "tool_use":
			var input map[string]any
			if err := json.Unmarshal(block.Input, &input); err != nil {
				return CompletionResponse{}, fmt.Errorf("unmarshal tool input for %q: %w", block.Name, err)
			}
			resp.ToolCalls = append(resp.ToolCalls, ToolCall{
				Name:  block.Name,
				Input: input,
			})
		}
	}
	resp.Content = sb.String()

	return resp, nil
}
