package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/option"
)

// Translator calls the LLM to translate an option's text fields into a target locale.
type Translator struct {
	router *llm.SmartRouter
}

// NewTranslator creates a Translator backed by the given SmartRouter.
func NewTranslator(router *llm.SmartRouter) *Translator {
	return &Translator{router: router}
}

// Translate translates the option's text fields into the given locale (e.g. "fr", "de").
// Returns a json.RawMessage containing a TranslationBlob ready to be stored.
func (t *Translator) Translate(ctx context.Context, o option.Option, locale string) (json.RawMessage, error) {
	input := TranslatableFields(o)

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("marshal translation input: %w", err)
	}

	systemPrompt := fmt.Sprintf(
		"You are a professional translator. Translate the JSON object into %s. "+
			"Return ONLY a single raw JSON object. Do NOT use markdown fences. Do NOT add any explanation. "+
			"The output must match the exact same schema as the input. "+
			"Do not add or remove fields. Translate all string values. "+
			"Keep the same array length for attributes and warnings.",
		locale,
	)

	ctx = llm.WithRequestOpts(ctx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "translate",
	})

	req := llm.CompletionRequest{
		Messages: []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: string(inputJSON)},
		},
		MaxTokens: 2048,
	}

	resp, err := t.router.Complete(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("translate to %s: %w", locale, err)
	}

	raw := strings.TrimSpace(resp.Content)
	raw = stripFences(raw)

	// Validate the response can be parsed as a TranslationBlob.
	var blob TranslationBlob
	if err := json.Unmarshal([]byte(raw), &blob); err != nil {
		return nil, fmt.Errorf("invalid translation response for %s: %w", locale, err)
	}

	return json.RawMessage(raw), nil
}

// stripFences removes leading/trailing markdown code fences if present.
func stripFences(s string) string {
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if idx := strings.Index(s, "\n"); idx != -1 {
		s = s[idx+1:]
	} else {
		return ""
	}
	if strings.HasSuffix(s, "```") {
		if idx := strings.LastIndex(s, "\n"); idx != -1 {
			s = s[:idx]
		} else {
			return ""
		}
	}
	return strings.TrimSpace(s)
}
