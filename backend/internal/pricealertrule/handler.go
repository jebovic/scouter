package pricealertrule

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
)

// Handler exposes HTTP endpoints for managing price alert rules.
type Handler struct {
	repo *Repository
}

// NewHandler returns a Handler backed by the provided Repository.
func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// createRuleRequest is the JSON body for POST /api/items/{itemID}/alert-rules.
type createRuleRequest struct {
	Type      RuleType `json:"type"`
	Value     float64  `json:"value"`
	BasePrice float64  `json:"basePrice"`
}

// checkRuleRequest is the JSON body for POST /api/items/{itemID}/alert-rules/check.
type checkRuleRequest struct {
	CurrentPrice float64  `json:"currentPrice"`
	WeeklyLow    *float64 `json:"weeklyLow"`
}

// CreateRule handles POST /api/items/{itemID}/alert-rules.
func (h *Handler) CreateRule(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "itemID is required")
		return
	}

	var req createRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !isValidRuleType(req.Type) {
		httputil.WriteError(w, http.StatusBadRequest, "type must be one of: drop_percent, fixed_below, weekly_low")
		return
	}

	if req.Type != RuleWeeklyLow && req.Value <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "value must be positive for this rule type")
		return
	}

	rule := AlertRule{
		ID:        uuid.NewString(),
		ItemID:    itemID,
		Type:      req.Type,
		Value:     req.Value,
		BasePrice: req.BasePrice,
		CreatedAt: time.Now().UTC(),
		Triggered: false,
	}

	if err := h.repo.AddRule(rule); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to add rule")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, rule)
}

// ListRules handles GET /api/items/{itemID}/alert-rules.
func (h *Handler) ListRules(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "itemID is required")
		return
	}

	rules := h.repo.ListByItem(itemID)
	if rules == nil {
		rules = []AlertRule{}
	}
	httputil.WriteJSON(w, http.StatusOK, rules)
}

// DeleteRule handles DELETE /api/items/{itemID}/alert-rules/{ruleID}.
func (h *Handler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID := chi.URLParam(r, "ruleID")
	if ruleID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "ruleID is required")
		return
	}

	if err := h.repo.DeleteRule(ruleID); err != nil {
		httputil.WriteError(w, http.StatusNotFound, "rule not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CheckRules handles POST /api/items/{itemID}/alert-rules/check.
func (h *Handler) CheckRules(w http.ResponseWriter, r *http.Request) {
	itemID := chi.URLParam(r, "itemID")
	if itemID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "itemID is required")
		return
	}

	var req checkRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.CurrentPrice < 0 {
		httputil.WriteError(w, http.StatusBadRequest, "currentPrice must be non-negative")
		return
	}

	rules := h.repo.ListByItem(itemID)
	results := make([]CheckResult, 0, len(rules))

	for _, rule := range rules {
		result := evaluateRule(rule, req.CurrentPrice, req.WeeklyLow)
		if result.Triggered {
			// Best-effort mark; ignore errors (e.g., concurrent delete).
			_ = h.repo.MarkTriggered(rule.ID)
		}
		results = append(results, result)
	}

	httputil.WriteJSON(w, http.StatusOK, results)
}

// evaluateRule checks whether the rule fires given the current price context.
func evaluateRule(rule AlertRule, currentPrice float64, weeklyLow *float64) CheckResult {
	res := CheckResult{
		ItemID: rule.ItemID,
		RuleID: rule.ID,
	}

	switch rule.Type {
	case RuleDropPercent:
		if rule.BasePrice <= 0 {
			res.Message = "base price not set; cannot evaluate drop_percent rule"
			return res
		}
		drop := (rule.BasePrice - currentPrice) / rule.BasePrice * 100
		if drop >= rule.Value {
			res.Triggered = true
			res.Message = fmt.Sprintf("Price dropped %.1f%% from %.2f to %.2f (threshold: %.1f%%)",
				drop, rule.BasePrice, currentPrice, rule.Value)
		} else {
			res.Message = fmt.Sprintf("Drop of %.1f%% not yet reached (need %.1f%%)", drop, rule.Value)
		}

	case RuleFixedBelow:
		if currentPrice < rule.Value {
			res.Triggered = true
			res.Message = fmt.Sprintf("Price %.2f is below threshold %.2f", currentPrice, rule.Value)
		} else {
			res.Message = fmt.Sprintf("Price %.2f is not below threshold %.2f", currentPrice, rule.Value)
		}

	case RuleWeeklyLow:
		if weeklyLow == nil {
			res.Message = "weeklyLow not provided; cannot evaluate weekly_low rule"
			return res
		}
		if currentPrice <= *weeklyLow {
			res.Triggered = true
			res.Message = fmt.Sprintf("Price %.2f matches or beats the weekly low of %.2f", currentPrice, *weeklyLow)
		} else {
			res.Message = fmt.Sprintf("Price %.2f is above the weekly low of %.2f", currentPrice, *weeklyLow)
		}

	default:
		res.Message = "unknown rule type"
	}

	return res
}

// isValidRuleType reports whether t is one of the recognised rule types.
func isValidRuleType(t RuleType) bool {
	switch t {
	case RuleDropPercent, RuleFixedBelow, RuleWeeklyLow:
		return true
	default:
		return false
	}
}
