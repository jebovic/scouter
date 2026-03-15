package negotiation

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// Handler serves the negotiation endpoint.
type Handler struct {
	agent       *Agent
	optionRepo  option.Repository
	missionRepo mission.Repository
	shoppingRepo shopping.Repository
}

// NewHandler creates a new negotiation Handler backed by a pgxpool.Pool.
// pool may be nil in tests that don't exercise DB paths.
func NewHandler(pool *pgxpool.Pool, provider llm.Provider) *Handler {
	h := &Handler{
		agent: NewAgent(provider),
	}
	if pool != nil {
		h.optionRepo = option.NewRepository(pool)
		h.missionRepo = mission.NewRepository(pool)
		h.shoppingRepo = shopping.NewRepository(pool)
	}
	return h
}

// NewHandlerWithRepos creates a Handler with pre-built repository stubs.
// Intended for unit tests that need full handler coverage without a live database.
func NewHandlerWithRepos(
	optRepo option.Repository,
	missionRepo mission.Repository,
	shoppingRepo shopping.Repository,
	provider llm.Provider,
) *Handler {
	return &Handler{
		agent:        NewAgent(provider),
		optionRepo:   optRepo,
		missionRepo:  missionRepo,
		shoppingRepo: shoppingRepo,
	}
}

// Routes returns a chi.Router mounted at /api/options/{optionID}.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/negotiate", h.negotiate)
	return r
}

// negotiate handles POST /api/options/{optionID}/negotiate.
func (h *Handler) negotiate(w http.ResponseWriter, r *http.Request) {
	optionIDStr := chi.URLParam(r, "optionID")
	optionID, err := uuid.Parse(optionIDStr)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid option ID")
		return
	}

	// Fetch option
	if h.optionRepo == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	opt, err := h.optionRepo.GetByID(r.Context(), optionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if opt == nil {
		httputil.WriteError(w, http.StatusNotFound, "option not found")
		return
	}

	// Fetch parent mission for context
	m, err := h.missionRepo.GetByID(r.Context(), opt.MissionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if m == nil {
		httputil.WriteError(w, http.StatusNotFound, "mission not found")
		return
	}

	// Optionally parse a request body for override hints (future-proofing)
	var override struct {
		Budget   *float64 `json:"budget,omitempty"`
		Currency *string  `json:"currency,omitempty"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		_ = json.NewDecoder(r.Body).Decode(&override)
	}

	// Build NegotiationInput from DB data
	input := buildInput(opt, m)
	if override.Budget != nil && *override.Budget > 0 {
		input.Budget = *override.Budget
	}
	if override.Currency != nil && *override.Currency != "" {
		input.Currency = *override.Currency
	}

	// Fetch market prices from shopping items (best-effort, graceful on error)
	if h.shoppingRepo != nil {
		input.MarketPrices = fetchMarketPrices(r.Context(), h.shoppingRepo, m.ID, opt.Name)
	}

	script, err := h.agent.Run(r.Context(), input)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("negotiation agent error: %v", err))
		return
	}

	httputil.WriteJSON(w, http.StatusOK, script)
}

// buildInput constructs a NegotiationInput from an option and its parent mission.
func buildInput(opt *option.Option, m *mission.Mission) NegotiationInput {
	// Convert option attributes to a simple map
	attrs := make(map[string]string, len(opt.Attributes))
	for _, a := range opt.Attributes {
		attrs[a.Key] = fmt.Sprintf("%v", a.Value)
	}

	// Convert mission constraints
	constraints := make([]ConstraintInfo, 0, len(m.Constraints))
	for _, c := range m.Constraints {
		constraints = append(constraints, ConstraintInfo{
			Key:   c.Key,
			Label: c.Label,
			Value: fmt.Sprintf("%v", c.Value),
			Type:  c.Type,
		})
	}

	// Use the best price from the option's price range as the option price
	optionPrice := 0.0
	if opt.PriceRange != nil {
		optionPrice = opt.PriceRange.Best
		if optionPrice <= 0 {
			optionPrice = opt.PriceRange.Min
		}
	}

	return NegotiationInput{
		MissionName: m.Name,
		Category:    m.Category,
		Budget:      m.Budget,
		Currency:    m.Currency,
		OptionName:  opt.Name,
		OptionPrice: optionPrice,
		Attributes:  attrs,
		Constraints: constraints,
	}
}

// fetchMarketPrices looks up shopping items for the same mission and collects
// merchant/price pairs that match the option name. Non-fatal — returns empty on error.
func fetchMarketPrices(ctx context.Context, repo shopping.Repository, missionID uuid.UUID, optionName string) []MarketPrice {
	items, err := repo.ListByMission(ctx, missionID)
	if err != nil {
		return nil
	}

	var prices []MarketPrice
	for _, item := range items {
		// Include items whose name contains the option name (case-insensitive prefix check)
		if containsIgnoreCase(item.Name, optionName) && item.Merchant != "" {
			prices = append(prices, MarketPrice{
				Merchant: item.Merchant,
				Price:    item.Price,
			})
		}
	}
	return prices
}

func containsIgnoreCase(s, substr string) bool {
	if substr == "" {
		return true
	}
	sl := len(s)
	subl := len(substr)
	if subl > sl {
		return false
	}
	for i := 0; i <= sl-subl; i++ {
		match := true
		for j := 0; j < subl; j++ {
			cs := s[i+j]
			ct := substr[j]
			if cs >= 'A' && cs <= 'Z' {
				cs += 32
			}
			if ct >= 'A' && ct <= 'Z' {
				ct += 32
			}
			if cs != ct {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

