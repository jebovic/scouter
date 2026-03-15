package summary

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// Generator is the narrow interface consumed by Handler so tests can inject
// a fake without depending on the real LLM.
type Generator interface {
	Generate(ctx context.Context, m mission.Mission, opts []option.Option, items []shopping.Item) (string, error)
}

// Agent uses the LLM to generate a French markdown shopping summary report.
// It uses text-only mode (no tool calls needed — just text generation).
type Agent struct {
	provider llm.Provider
}

// NewAgent creates a new SummaryAgent backed by the given LLM provider.
func NewAgent(provider llm.Provider) *Agent {
	return &Agent{provider: provider}
}

// Generate calls the LLM and returns the markdown summary text.
func (a *Agent) Generate(ctx context.Context, m mission.Mission, opts []option.Option, items []shopping.Item) (string, error) {
	req := a.buildRequest(m, opts, items)

	llmCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	llmCtx = llm.WithRequestOpts(llmCtx, llm.RequestOpts{
		Capabilities: llm.CapText,
		Label:        "summary",
	})

	resp, err := a.provider.Complete(llmCtx, req)
	if err != nil {
		return "", fmt.Errorf("summary llm call: %w", err)
	}

	content := strings.TrimSpace(resp.Content)
	if content == "" {
		return "", fmt.Errorf("summary: LLM returned empty content")
	}
	return content, nil
}

func (a *Agent) buildRequest(m mission.Mission, opts []option.Option, items []shopping.Item) llm.CompletionRequest {
	prompt := buildPrompt(m, opts, items)
	return llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		MaxTokens: 1500,
	}
}

func buildPrompt(m mission.Mission, opts []option.Option, items []shopping.Item) string {
	var sb strings.Builder

	sb.WriteString("Tu es un conseiller d'achat français expert. ")
	sb.WriteString("Génère un rapport de synthèse d'achat concis en markdown français pour cette mission.\n\n")

	// Mission context
	sb.WriteString("## Mission\n")
	fmt.Fprintf(&sb, "- **Nom**: %s\n", m.Name)
	fmt.Fprintf(&sb, "- **Catégorie**: %s\n", m.Category)
	fmt.Fprintf(&sb, "- **Budget**: %.2f %s\n\n", m.Budget, m.Currency)

	// Options
	if len(opts) > 0 {
		sb.WriteString("## Options recherchées\n")
		for _, o := range opts {
			priceStr := ""
			if o.PriceRange != nil {
				priceStr = fmt.Sprintf(" — %.2f %s", o.PriceRange.Best, m.Currency)
			}
			fmt.Fprintf(&sb, "- **%s** (%s)%s\n", o.Name, o.Badge, priceStr)
		}
		sb.WriteString("\n")
	}

	// Shopping items
	if len(items) > 0 {
		sb.WriteString("## Articles de shopping\n")
		for _, i := range items {
			fmt.Fprintf(&sb, "- **%s** chez %s : %.2f %s (%s)\n", i.Name, i.Merchant, i.Price, m.Currency, i.Status)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Instructions\n")
	sb.WriteString("Rédige un rapport structuré en markdown français avec exactement ces sections:\n")
	sb.WriteString("1. **Résumé exécutif** (2-3 phrases)\n")
	sb.WriteString("2. **Top 3 options recommandées** avec justification courte\n")
	sb.WriteString("3. **Meilleur prix trouvé** (marchand et montant)\n")
	sb.WriteString("4. **Conseil sur le moment d'achat** (basé sur les soldes françaises)\n")
	sb.WriteString("5. **Évaluation budgétaire** (le budget est-il suffisant ?)\n\n")
	sb.WriteString("Sois concis. Maximum 500 mots. Utilise le markdown (##, **gras**, listes).\n")

	return sb.String()
}
