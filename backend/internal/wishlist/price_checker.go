package wishlist

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/notification"
)

const checkInterval = 24 * time.Hour

// PriceChecker periodically estimates market prices for wish list items that
// have a target_price set and alert_enabled=true.
type PriceChecker struct {
	repo      *Repository
	notifRepo notification.Repository
	provider  llm.Provider
	log       *slog.Logger
}

// NewPriceChecker creates a PriceChecker. log may be nil (no-op logger used).
func NewPriceChecker(repo *Repository, notifRepo notification.Repository, provider llm.Provider) *PriceChecker {
	log := slog.Default()
	return &PriceChecker{
		repo:      repo,
		notifRepo: notifRepo,
		provider:  provider,
		log:       log,
	}
}

// CheckAll queries all eligible wish list items and runs a price check for each.
func (c *PriceChecker) CheckAll(ctx context.Context) error {
	items, err := c.repo.ListAlertCandidates(ctx)
	if err != nil {
		return fmt.Errorf("wishlist price checker: list candidates: %w", err)
	}

	for _, item := range items {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := c.checkItem(ctx, item); err != nil {
			c.log.Warn("wishlist price checker: item check failed",
				"item_id", item.ID, "name", item.Name, "err", err)
		}
	}
	return nil
}

// checkItem runs a single LLM price estimate and records/notifies as needed.
func (c *PriceChecker) checkItem(ctx context.Context, item WishListItem) error {
	prompt := buildPricePrompt(item)
	tool := estimatePriceTool()

	resp, err := c.provider.Complete(ctx, llm.CompletionRequest{
		Messages:  []llm.Message{{Role: "user", Content: prompt}},
		Tools:     []llm.Tool{tool},
		MaxTokens: 256,
	})
	if err != nil {
		return fmt.Errorf("llm complete: %w", err)
	}

	estimatedPrice, err := parseEstimatedPrice(resp.ToolCalls)
	if err != nil {
		return fmt.Errorf("parse estimate: %w", err)
	}

	if err := c.repo.UpdatePriceCheck(ctx, item.ID, estimatedPrice); err != nil {
		return fmt.Errorf("update price check: %w", err)
	}

	if item.TargetPrice != nil && estimatedPrice <= *item.TargetPrice {
		currency := item.Currency
		title := fmt.Sprintf("Price alert: %s", item.Name)
		body := fmt.Sprintf("Estimated at %.2f %s (target: %.2f %s).",
			estimatedPrice, currency, *item.TargetPrice, currency)

		if _, err := c.notifRepo.Create(ctx, notification.CreateRequest{
			Type:  notification.TypeTargetHit,
			Title: title,
			Body:  body,
		}); err != nil {
			return fmt.Errorf("create notification: %w", err)
		}
	}

	return nil
}

// buildPricePrompt constructs the user prompt for the LLM estimate.
func buildPricePrompt(item WishListItem) string {
	url := "(none)"
	if item.URL != nil {
		url = *item.URL
	}
	notes := "(none)"
	if item.Notes != nil {
		notes = *item.Notes
	}
	return fmt.Sprintf(
		"You are a price intelligence expert. "+
			"Estimate the current market price for: %s in %s. "+
			"Notes: %s. URL: %s.",
		item.Name, item.Currency, notes, url,
	)
}

// estimatePriceTool returns the tool schema the LLM should call.
func estimatePriceTool() llm.Tool {
	return llm.Tool{
		Name:        "estimate_price",
		Description: "Estimate the current market price for a product",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"estimated_price": map[string]any{
					"type":        "number",
					"description": "Estimated current price in the given currency",
				},
				"currency": map[string]any{
					"type": "string",
				},
				"confidence": map[string]any{
					"type": "string",
					"enum": []string{"low", "medium", "high"},
				},
			},
			"required": []string{"estimated_price", "currency", "confidence"},
		},
	}
}

// parseEstimatedPrice extracts the estimated_price value from a tool call response.
func parseEstimatedPrice(calls []llm.ToolCall) (float64, error) {
	for _, call := range calls {
		if call.Name != "estimate_price" {
			continue
		}
		val, ok := call.Input["estimated_price"]
		if !ok {
			return 0, fmt.Errorf("estimate_price tool call missing estimated_price field")
		}
		price, ok := val.(float64)
		if !ok {
			return 0, fmt.Errorf("estimated_price is not a number: %T", val)
		}
		return price, nil
	}
	return 0, fmt.Errorf("no estimate_price tool call in response")
}

// shouldCheck reports whether an item is eligible for a price check.
// It returns false if alert is disabled, target price is nil, or the item was
// checked within the last 24 hours.
func shouldCheck(item WishListItem) bool {
	if !item.AlertEnabled || item.TargetPrice == nil {
		return false
	}
	if item.LastCheckedAt == nil {
		return true
	}
	return time.Since(*item.LastCheckedAt) >= checkInterval
}
