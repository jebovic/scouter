package substitute_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jibei/scouter/internal/substitute"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeSubstituteAgent struct {
	result []substitute.Substitute
	err    error
	called int
}

func (f *fakeSubstituteAgent) FindSubstitutes(_ context.Context, name, category string, budget float64) ([]substitute.Substitute, error) {
	f.called++
	return f.result, f.err
}

type fakeOptionGetter struct {
	name     string
	category string
	budget   float64
	err      error
}

func (f *fakeOptionGetter) GetOptionInfo(_ context.Context, optionID string) (name, category string, budget float64, err error) {
	return f.name, f.category, f.budget, f.err
}

// ── router helper ────────────────────────────────────────────────────────────

func mountSubstituteRouter(h *substitute.Handler) http.Handler {
	r := chi.NewRouter()
	r.Get("/api/options/{id}/substitutes", h.GetSubstitutes)
	return r
}

func goodSubstituteResponse(optionID string) *substitute.SubstituteResponse {
	return &substitute.SubstituteResponse{
		OptionID:    optionID,
		ProductName: "iPhone 15 Pro",
		Substitutes: sampleSubstitutes(),
		GeneratedAt: time.Now(),
	}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestGetSubstitutes_HappyPath(t *testing.T) {
	const optID = "11111111-1111-1111-1111-111111111111"
	agent := &fakeSubstituteAgent{result: sampleSubstitutes()}
	getter := &fakeOptionGetter{name: "iPhone 15 Pro", category: "smartphone", budget: 1000.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got substitute.SubstituteResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.OptionID != optID {
		t.Errorf("expected optionId %s, got %s", optID, got.OptionID)
	}
	if len(got.Substitutes) != 2 {
		t.Errorf("expected 2 substitutes, got %d", len(got.Substitutes))
	}
	if agent.called != 1 {
		t.Errorf("expected agent called once, got %d", agent.called)
	}
}

// TestGetSubstitutes_ThreeSubstitutes verifies the happy path returns exactly 3
// substitutes when the agent provides them (Phase 79: "success 3 substitutes").
func TestGetSubstitutes_ThreeSubstitutes(t *testing.T) {
	const optID = "11111111-1111-1111-1111-111111111112"
	agent := &fakeSubstituteAgent{result: threeSubstitutes()}
	getter := &fakeOptionGetter{name: "Sony WH-1000XM5", category: "audio", budget: 350.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got substitute.SubstituteResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got.Substitutes) != 3 {
		t.Errorf("expected 3 substitutes, got %d", len(got.Substitutes))
	}
}

func TestGetSubstitutes_CacheHit(t *testing.T) {
	const optID = "22222222-2222-2222-2222-222222222222"
	agent := &fakeSubstituteAgent{result: sampleSubstitutes()}
	getter := &fakeOptionGetter{name: "iPhone 15 Pro", category: "smartphone", budget: 1000.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)
	router := mountSubstituteRouter(h)

	// First request — populates cache.
	r1 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request failed: %d", w1.Code)
	}

	// Second request — must be served from cache.
	r2 := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, r2)
	if w2.Code != http.StatusOK {
		t.Fatalf("second request failed: %d", w2.Code)
	}

	if agent.called != 1 {
		t.Errorf("expected agent called exactly once (cache hit on 2nd), got %d", agent.called)
	}
}

func TestGetSubstitutes_OptionNotFound(t *testing.T) {
	const optID = "33333333-3333-3333-3333-333333333333"
	agent := &fakeSubstituteAgent{}
	getter := &fakeOptionGetter{name: ""} // empty name → not found
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when option not found, called %d times", agent.called)
	}
}

// TestGetSubstitutes_SavingsPercentPositive verifies that substitutes with
// SavingsPercent are surfaced correctly and are non-negative (Phase 79).
func TestGetSubstitutes_SavingsPercentPositive(t *testing.T) {
	const optID = "44444444-4444-4444-4444-444444444444"
	subs := substituteWithSavings()
	agent := &fakeSubstituteAgent{result: subs}
	getter := &fakeOptionGetter{name: "Bose QC45", category: "audio", budget: 329.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got substitute.SubstituteResponse
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for i, s := range got.Substitutes {
		if s.SavingsPercent < 0 {
			t.Errorf("substitute[%d] has negative SavingsPercent: %.2f", i, s.SavingsPercent)
		}
	}
	// At least one substitute should have a positive savings percentage.
	found := false
	for _, s := range got.Substitutes {
		if s.SavingsPercent > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected at least one substitute with SavingsPercent > 0")
	}
}

// TestGetSubstitutes_CountInRange verifies the Phase 79 contract: substitutes
// count must be in [3, 5] when the agent returns that many.
func TestGetSubstitutes_CountInRange(t *testing.T) {
	for _, n := range []int{3, 4, 5} {
		subs := nSubstitutes(n)
		agent := &fakeSubstituteAgent{result: subs}
		getter := &fakeOptionGetter{name: "Product", category: "electronics", budget: 100.0}
		cache := substitute.NewCache(2 * time.Hour)
		h := substitute.NewHandler(getter, cache, agent)

		optID := "55555555-5555-5555-5555-55555555555" + string(rune('0'+n))
		req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
		w := httptest.NewRecorder()
		mountSubstituteRouter(h).ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("n=%d: expected 200, got %d: %s", n, w.Code, w.Body.String())
		}
		var got substitute.SubstituteResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("n=%d: decode: %v", n, err)
		}
		if len(got.Substitutes) != n {
			t.Errorf("n=%d: expected %d substitutes, got %d", n, n, len(got.Substitutes))
		}
		count := len(got.Substitutes)
		if count < 3 || count > 5 {
			t.Errorf("n=%d: substitutes count %d out of [3,5] range", n, count)
		}
	}
}

func TestGetSubstitutes_GetterError(t *testing.T) {
	const optID = "66666666-6666-6666-6666-666666666666"
	agent := &fakeSubstituteAgent{}
	getter := &fakeOptionGetter{err: errors.New("db error")}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
	if agent.called != 0 {
		t.Errorf("agent should not be called when getter fails, called %d times", agent.called)
	}
}

func TestGetSubstitutes_AgentError(t *testing.T) {
	const optID = "77777777-7777-7777-7777-777777777777"
	agent := &fakeSubstituteAgent{err: errors.New("llm unavailable")}
	getter := &fakeOptionGetter{name: "Sony TV", category: "electronics", budget: 800.0}
	cache := substitute.NewCache(2 * time.Hour)
	h := substitute.NewHandler(getter, cache, agent)

	req := httptest.NewRequest(http.MethodGet, "/api/options/"+optID+"/substitutes", nil)
	w := httptest.NewRecorder()
	mountSubstituteRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d: %s", w.Code, w.Body.String())
	}
}

// ── helper constructors ───────────────────────────────────────────────────────

func threeSubstitutes() []substitute.Substitute {
	return []substitute.Substitute{
		{
			Name: "Anker Soundcore Q45", Brand: "Anker", Price: 59.99,
			Currency: "EUR", Retailer: "Amazon France",
			Reason: "Excellent rapport qualité-prix", Advantage: "cheaper",
			SavingsPercent: 46,
			Pros:           []string{"Autonomie 60h", "ANC efficace", "Léger"},
			Cons:           []string{"Son moins précis", "Pas de multipoint"},
			WhyConsider:    "La meilleure alternative budget pour 90% des usages quotidiens.",
		},
		{
			Name: "JBL Tune 760NC", Brand: "JBL", Price: 89.99,
			Currency: "EUR", Retailer: "Fnac",
			Reason: "Bonne réduction de bruit pour le prix", Advantage: "cheaper",
			SavingsPercent: 27,
			Pros:           []string{"Basses puissantes", "50h d'autonomie"},
			Cons:           []string{"Micro médiocre"},
			WhyConsider:    "Idéal pour les amateurs de basses.",
		},
		{
			Name: "Sony WH-CH720N", Brand: "Sony", Price: 129.99,
			Currency: "EUR", Retailer: "Darty",
			Reason: "Meilleur ANC dans cette gamme de prix", Advantage: "better_rated",
			SavingsPercent: 8,
			Pros:           []string{"ANC top", "Léger 192g", "Appel multipoint"},
			Cons:           []string{"Basses légères"},
			WhyConsider:    "Le meilleur choix si l'ANC est prioritaire.",
		},
	}
}

func substituteWithSavings() []substitute.Substitute {
	return []substitute.Substitute{
		{
			Name: "Jabra Evolve2 30", Brand: "Jabra", Price: 189.0,
			Currency: "EUR", Retailer: "Boulanger",
			Reason: "Qualité professionnelle à moindre coût", Advantage: "cheaper",
			SavingsPercent: 43,
			Pros:           []string{"Micro certifié Teams"},
			Cons:           []string{"Pas d'ANC"},
			WhyConsider:    "Parfait pour les visioconférences.",
		},
		{
			Name: "Jabra Evolve2 55", Brand: "Jabra", Price: 280.0,
			Currency: "EUR", Retailer: "Cdiscount",
			Reason: "ANC avancé, idéal en open space", Advantage: "better_rated",
			SavingsPercent: 15,
			Pros:           []string{"ANC avancé", "8 micros"},
			Cons:           []string{"Lourd"},
			WhyConsider:    "La référence bureautique avec ANC.",
		},
	}
}

func nSubstitutes(n int) []substitute.Substitute {
	base := substitute.Substitute{
		Brand: "Brand", Currency: "EUR", Retailer: "Fnac",
		Reason: "Good substitute", Advantage: "cheaper",
		SavingsPercent: 20,
		Pros:           []string{"Bon rapport qualité-prix"},
		Cons:           []string{"Moins connu"},
		WhyConsider:    "Une bonne alternative.",
	}
	result := make([]substitute.Substitute, n)
	for i := range result {
		s := base
		s.Name = "Product " + string(rune('A'+i))
		s.Price = float64(50 + i*10)
		result[i] = s
	}
	return result
}
