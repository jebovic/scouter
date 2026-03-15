package health

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
)

// Summarizer is the narrow interface consumed by Handler to allow test injection.
type Summarizer interface {
	BuildHealthSummary(ctx context.Context, missionName string, report *HealthReport) (summary, advice string, err error)
}

// Agent uses the LLM to generate a French summary and advice for the health report.
// Only text-only mode is used (CapText) — no tool calls.
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new health Agent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// BuildHealthSummary calls the LLM with a short prompt and returns a French
// 1–2 sentence summary and 1 sentence actionable advice.
func (a *Agent) BuildHealthSummary(ctx context.Context, missionName string, report *HealthReport) (string, string, error) {
	prompt := buildSummaryPrompt(missionName, report)

	llmCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "health-summary",
	})

	resp, err := a.provider.Complete(llmCtx, llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		MaxTokens: 300,
	})
	if err != nil {
		return "", "", fmt.Errorf("health summary llm call: %w", err)
	}

	return parseSummaryResponse(resp.Content)
}

// buildSummaryPrompt constructs the prompt sent to the LLM.
func buildSummaryPrompt(missionName string, report *HealthReport) string {
	var sb strings.Builder
	sb.WriteString("Tu es un conseiller d'achat français expert. ")
	sb.WriteString("Analyse la santé de cette mission d'achat et réponds en français uniquement.\n\n")

	fmt.Fprintf(&sb, "Mission: %s\n", missionName)
	fmt.Fprintf(&sb, "Score global: %d/100 (Grade: %s)\n", report.OverallScore, report.Grade)
	sb.WriteString("Signaux:\n")
	for _, s := range report.Signals {
		fmt.Fprintf(&sb, "- %s %s: %d/100 (poids: %.0f%%)\n", s.Icon, s.Name, s.Score, s.Weight*100)
	}

	sb.WriteString("\nRéponds EXACTEMENT avec ce format (2 lignes):\n")
	sb.WriteString("RÉSUMÉ: [1-2 phrases résumant la santé de la mission]\n")
	sb.WriteString("CONSEIL: [1 phrase de conseil actionnable]\n")

	return sb.String()
}

// parseSummaryResponse extracts the RÉSUMÉ and CONSEIL from the LLM response.
// If parsing fails, it returns the full content as summary with a default advice.
func parseSummaryResponse(content string) (string, string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return "", "", fmt.Errorf("health summary: LLM returned empty content")
	}

	var summary, advice string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "RÉSUMÉ:"); ok {
			summary = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "CONSEIL:"); ok {
			advice = strings.TrimSpace(after)
		}
	}

	if summary == "" {
		summary = content
	}
	if advice == "" {
		advice = "Continuez à améliorer votre mission pour atteindre un score optimal."
	}

	return summary, advice, nil
}
