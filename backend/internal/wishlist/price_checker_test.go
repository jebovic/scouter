package wishlist

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/notification"
)

var _ notification.Repository = (*stubNotifRepo)(nil) // compile-time interface check

// -- stubs --

type stubProvider struct {
	resp llm.CompletionResponse
	err  error
}

func (s *stubProvider) Complete(_ context.Context, _ llm.CompletionRequest) (llm.CompletionResponse, error) {
	return s.resp, s.err
}

type stubNotifRepo struct {
	created []notification.CreateRequest
	err     error
}

func (s *stubNotifRepo) List(_ context.Context, _ int) ([]notification.Notification, error) {
	return nil, nil
}
func (s *stubNotifRepo) UnreadCount(_ context.Context) (int, error) { return 0, nil }
func (s *stubNotifRepo) Create(_ context.Context, req notification.CreateRequest) (*notification.Notification, error) {
	s.created = append(s.created, req)
	return nil, s.err
}
func (s *stubNotifRepo) MarkRead(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubNotifRepo) MarkAllRead(_ context.Context) error { return nil }

// -- tests --

func TestNewPriceChecker(t *testing.T) {
	repo := &Repository{}
	notifRepo := &stubNotifRepo{}
	provider := &stubProvider{}

	pc := NewPriceChecker(repo, notifRepo, provider)
	if pc == nil {
		t.Fatal("NewPriceChecker returned nil")
	}
	if pc.repo != repo {
		t.Error("repo not set correctly")
	}
	if pc.provider != provider {
		t.Error("provider not set correctly")
	}
}

func TestPriceChecker_buildPrompt(t *testing.T) {
	item := WishListItem{
		Name:     "Sony WH-1000XM5",
		Currency: "EUR",
	}
	prompt := buildPricePrompt(item)
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
	if len(prompt) < 10 {
		t.Error("prompt too short")
	}
}

func TestPriceChecker_buildPrompt_withURLAndNotes(t *testing.T) {
	url := "https://amazon.de/dp/B09XS7JWHH"
	notes := "Black Friday deal expected"
	item := WishListItem{
		Name:     "Sony WH-1000XM5",
		Currency: "USD",
		URL:      &url,
		Notes:    &notes,
	}
	prompt := buildPricePrompt(item)
	if prompt == "" {
		t.Error("prompt should not be empty")
	}
}

func TestPriceChecker_estimatePriceToolSchema(t *testing.T) {
	schema := estimatePriceTool()
	if schema.Name != "estimate_price" {
		t.Errorf("unexpected tool name: %s", schema.Name)
	}
	if schema.Description == "" {
		t.Error("tool description should not be empty")
	}
	if schema.InputSchema == nil {
		t.Error("input schema should not be nil")
	}
}

func TestPriceChecker_parseEstimatedPrice_valid(t *testing.T) {
	calls := []llm.ToolCall{
		{
			Name: "estimate_price",
			Input: map[string]any{
				"estimated_price": float64(249.99),
				"currency":        "EUR",
				"confidence":      "high",
			},
		},
	}
	price, err := parseEstimatedPrice(calls)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 249.99 {
		t.Errorf("expected 249.99, got %f", price)
	}
}

func TestPriceChecker_parseEstimatedPrice_noToolCalls(t *testing.T) {
	_, err := parseEstimatedPrice(nil)
	if err == nil {
		t.Error("expected error for empty tool calls")
	}
}

func TestPriceChecker_parseEstimatedPrice_wrongTool(t *testing.T) {
	calls := []llm.ToolCall{
		{Name: "other_tool", Input: map[string]any{"foo": "bar"}},
	}
	_, err := parseEstimatedPrice(calls)
	if err == nil {
		t.Error("expected error when no estimate_price tool call found")
	}
}

func TestPriceChecker_parseEstimatedPrice_missingField(t *testing.T) {
	calls := []llm.ToolCall{
		{Name: "estimate_price", Input: map[string]any{}},
	}
	_, err := parseEstimatedPrice(calls)
	if err == nil {
		t.Error("expected error when estimated_price field missing")
	}
}

func TestPriceChecker_shouldCheck_noTargetPrice(t *testing.T) {
	item := WishListItem{
		AlertEnabled: true,
		TargetPrice:  nil,
	}
	if shouldCheck(item) {
		t.Error("should not check item without target price")
	}
}

func TestPriceChecker_shouldCheck_alertDisabled(t *testing.T) {
	price := 100.0
	item := WishListItem{
		AlertEnabled: false,
		TargetPrice:  &price,
	}
	if shouldCheck(item) {
		t.Error("should not check item with alert disabled")
	}
}

func TestPriceChecker_shouldCheck_recentlyChecked(t *testing.T) {
	price := 100.0
	recent := time.Now().Add(-12 * time.Hour)
	item := WishListItem{
		AlertEnabled:  true,
		TargetPrice:   &price,
		LastCheckedAt: &recent,
	}
	if shouldCheck(item) {
		t.Error("should not check item checked less than 24h ago")
	}
}

func TestPriceChecker_shouldCheck_staleCheck(t *testing.T) {
	price := 100.0
	old := time.Now().Add(-25 * time.Hour)
	item := WishListItem{
		AlertEnabled:  true,
		TargetPrice:   &price,
		LastCheckedAt: &old,
	}
	if !shouldCheck(item) {
		t.Error("should check item not checked for more than 24h")
	}
}

func TestPriceChecker_shouldCheck_neverChecked(t *testing.T) {
	price := 100.0
	item := WishListItem{
		AlertEnabled:  true,
		TargetPrice:   &price,
		LastCheckedAt: nil,
	}
	if !shouldCheck(item) {
		t.Error("should check item that was never checked")
	}
}

func TestPriceChecker_LLMError(t *testing.T) {
	pc := &PriceChecker{
		provider: &stubProvider{err: errors.New("llm unreachable")},
	}
	item := WishListItem{ID: "id1", Name: "Test", Currency: "EUR"}
	err := pc.checkItem(context.Background(), item)
	if err == nil {
		t.Error("expected error when LLM call fails")
	}
}
