package negotiate_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/negotiate"
	"github.com/jibei/scouter/internal/shopping"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type fakeItemGetter struct {
	item *shopping.Item
	err  error
}

func (f *fakeItemGetter) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return f.item, f.err
}

type fakeNegotiateAgent struct {
	advice *negotiate.NegotiationAdvice
	err    error
	called int
}

func (f *fakeNegotiateAgent) Negotiate(_ context.Context, itemID, itemName, merchant string, currentPrice float64) (*negotiate.NegotiationAdvice, error) {
	f.called++
	return f.advice, f.err
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mountRouter(h *negotiate.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/shopping-items/{id}/negotiate", h.Get)
	return r
}

func sampleItem() *shopping.Item {
	return &shopping.Item{
		ID:       uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		Name:     "iPhone 15 Pro",
		Merchant: "Fnac",
		Price:    1299.0,
	}
}

func sampleAdvice(itemID string) *negotiate.NegotiationAdvice {
	return &negotiate.NegotiationAdvice{
		ItemID:       itemID,
		ItemName:     "iPhone 15 Pro",
		CurrentPrice: 1299.0,
		TargetPrice:  1130.13, // 1299 * 0.87
		Tips: []negotiate.NegotiationTip{
			{
				Title:       "Demander le prix affiché en magasin",
				Description: "Montrez les offres concurrentes sur votre téléphone et demandez poliment l'alignement de prix.",
				Difficulty:  "easy",
			},
			{
				Title:       "Négocier en fin de mois",
				Description: "Les vendeurs ont des objectifs mensuels — ils sont plus flexibles les derniers jours du mois.",
				Difficulty:  "easy",
			},
			{
				Title:       "Demander un bundle avec accessoires",
				Description: "Proposez d'acheter une coque ou des écouteurs en échange d'une remise sur le téléphone.",
				Difficulty:  "medium",
			},
			{
				Title:       "Programme de fidélité",
				Description: "Renseignez-vous sur les points fidélité ou les offres réservées aux membres Fnac+.",
				Difficulty:  "easy",
			},
		},
		Script:   "Bonjour, j'ai vu ce produit à 1129€ chez [concurrent]. Pouvez-vous vous aligner sur ce prix ?",
		CachedAt: time.Now().Unix(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

// TestNegotiate_Success verifies the happy path: item found, agent returns advice.
func TestNegotiate_Success(t *testing.T) {
	const itemID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	agent := &fakeNegotiateAgent{advice: sampleAdvice(itemID)}
	getter := &fakeItemGetter{item: sampleItem()}
	cache := negotiate.NewCache(24 * time.Hour)
	h := negotiate.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/negotiate", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got negotiate.NegotiationAdvice
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ItemID != itemID {
		t.Errorf("expected itemId %s, got %s", itemID, got.ItemID)
	}
	if got.ItemName != "iPhone 15 Pro" {
		t.Errorf("expected itemName 'iPhone 15 Pro', got %s", got.ItemName)
	}
	if len(got.Tips) == 0 {
		t.Error("expected at least one tip")
	}
	if got.Script == "" {
		t.Error("expected non-empty script")
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

// TestNegotiate_CacheHit verifies the second request is served from cache
// without calling the agent again.
func TestNegotiate_CacheHit(t *testing.T) {
	const itemID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	item := &shopping.Item{
		ID:       uuid.MustParse(itemID),
		Name:     "Sony WH-1000XM5",
		Merchant: "Darty",
		Price:    279.99,
	}
	advice := sampleAdvice(itemID)
	advice.ItemID = itemID
	advice.ItemName = item.Name
	advice.CurrentPrice = item.Price

	agent := &fakeNegotiateAgent{advice: advice}
	getter := &fakeItemGetter{item: item}
	cache := negotiate.NewCache(24 * time.Hour)
	h := negotiate.NewHandler(getter, agent, cache)
	router := mountRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/negotiate", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d %s", w1.Code, w1.Body.String())
	}

	// Second request — must be served from cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/negotiate", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called once (cache hit on 2nd), got %d", agent.called)
	}
}

// TestNegotiate_NotFound verifies 404 is returned when item does not exist.
func TestNegotiate_NotFound(t *testing.T) {
	const itemID = "cccccccc-cccc-cccc-cccc-cccccccccccc"

	agent := &fakeNegotiateAgent{}
	getter := &fakeItemGetter{item: nil, err: nil}
	cache := negotiate.NewCache(24 * time.Hour)
	h := negotiate.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/negotiate", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when item not found, called %d times", agent.called)
	}
}

// TestNegotiate_TargetPrice verifies the TargetPrice is computed as 13% below current price.
func TestNegotiate_TargetPrice(t *testing.T) {
	const itemID = "dddddddd-dddd-dddd-dddd-dddddddddddd"
	const price = 300.0
	// 300 * 0.87 = 261.00
	const expectedTarget = 261.0

	item := &shopping.Item{
		ID:       uuid.MustParse(itemID),
		Name:     "MacBook Air",
		Merchant: "Apple Store",
		Price:    price,
	}
	advice := &negotiate.NegotiationAdvice{
		ItemID:       itemID,
		ItemName:     item.Name,
		CurrentPrice: price,
		TargetPrice:  expectedTarget,
		Tips: []negotiate.NegotiationTip{
			{Title: "Tip1", Description: "Desc1", Difficulty: "easy"},
			{Title: "Tip2", Description: "Desc2", Difficulty: "medium"},
			{Title: "Tip3", Description: "Desc3", Difficulty: "hard"},
			{Title: "Tip4", Description: "Desc4", Difficulty: "easy"},
		},
		Script:   "Bonjour, je voudrais négocier ce prix.",
		CachedAt: time.Now().Unix(),
	}

	agent := &fakeNegotiateAgent{advice: advice}
	getter := &fakeItemGetter{item: item}
	cache := negotiate.NewCache(24 * time.Hour)
	h := negotiate.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/negotiate", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got negotiate.NegotiationAdvice
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.TargetPrice != expectedTarget {
		t.Errorf("expected targetPrice %.2f (13%% below %.2f), got %.2f", expectedTarget, price, got.TargetPrice)
	}
}

// TestNegotiate_TipsCount verifies the response contains between 4 and 6 tips.
func TestNegotiate_TipsCount(t *testing.T) {
	const itemID = "eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"

	item := &shopping.Item{
		ID:       uuid.MustParse(itemID),
		Name:     "Samsung TV 55",
		Merchant: "Boulanger",
		Price:    799.0,
	}
	advice := &negotiate.NegotiationAdvice{
		ItemID:       itemID,
		ItemName:     item.Name,
		CurrentPrice: item.Price,
		TargetPrice:  item.Price * 0.87,
		Tips: []negotiate.NegotiationTip{
			{Title: "T1", Description: "D1", Difficulty: "easy"},
			{Title: "T2", Description: "D2", Difficulty: "easy"},
			{Title: "T3", Description: "D3", Difficulty: "medium"},
			{Title: "T4", Description: "D4", Difficulty: "medium"},
			{Title: "T5", Description: "D5", Difficulty: "hard"},
		},
		Script:   "Bonjour, pouvez-vous faire un geste sur le prix ?",
		CachedAt: time.Now().Unix(),
	}

	agent := &fakeNegotiateAgent{advice: advice}
	getter := &fakeItemGetter{item: item}
	cache := negotiate.NewCache(24 * time.Hour)
	h := negotiate.NewHandler(getter, agent, cache)

	req := httptest.NewRequest(http.MethodGet, "/api/shopping-items/"+itemID+"/negotiate", nil)
	w := httptest.NewRecorder()
	mountRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got negotiate.NegotiationAdvice
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	count := len(got.Tips)
	if count < 4 || count > 6 {
		t.Errorf("expected 4-6 tips, got %d", count)
	}
}

// errAgentDown is a typed test error.
var errAgentDown = errors.New("agent unavailable")
