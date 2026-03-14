// Package export assembles and serializes a full mission report in JSON, Markdown, or PDF.
package export

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// ExportData is the full snapshot of a mission used for all export formats.
type ExportData struct {
	ExportedAt   time.Time                        `json:"exportedAt"`
	Mission      *mission.Mission                 `json:"mission"`
	Options      []option.Option                  `json:"options"`
	ShoppingItems []shopping.Item                 `json:"shoppingItems"`
	PriceHistory map[uuid.UUID][]shopping.PriceSnapshot `json:"priceHistory"`
	Decision     *decision.Decision               `json:"decision"`
}

// Gatherer assembles ExportData from the underlying repositories.
type Gatherer struct {
	missionRepo  mission.Repository
	optionRepo   option.Repository
	shoppingRepo shopping.Repository
	decisionRepo decision.Repository
}

// NewGatherer creates a new export Gatherer.
func NewGatherer(
	missionRepo mission.Repository,
	optionRepo option.Repository,
	shoppingRepo shopping.Repository,
	decisionRepo decision.Repository,
) *Gatherer {
	return &Gatherer{
		missionRepo:  missionRepo,
		optionRepo:   optionRepo,
		shoppingRepo: shoppingRepo,
		decisionRepo: decisionRepo,
	}
}

// Gather fetches all mission data in parallel-safe sequential calls.
func (g *Gatherer) Gather(ctx context.Context, missionID uuid.UUID) (*ExportData, error) {
	m, err := g.missionRepo.GetByID(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("export gather: mission: %w", err)
	}
	if m == nil {
		return nil, nil
	}

	opts, err := g.optionRepo.ListByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("export gather: options: %w", err)
	}
	if opts == nil {
		opts = []option.Option{}
	}

	items, err := g.shoppingRepo.ListByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("export gather: shopping: %w", err)
	}
	if items == nil {
		items = []shopping.Item{}
	}

	priceHistory := make(map[uuid.UUID][]shopping.PriceSnapshot)
	for _, item := range items {
		snapshots, err := g.shoppingRepo.ListPriceHistory(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("export gather: price history for %s: %w", item.ID, err)
		}
		if len(snapshots) > 0 {
			priceHistory[item.ID] = snapshots
		}
	}

	dec, err := g.decisionRepo.GetLatestByMission(ctx, missionID)
	if err != nil {
		return nil, fmt.Errorf("export gather: decision: %w", err)
	}

	return &ExportData{
		ExportedAt:    time.Now().UTC(),
		Mission:       m,
		Options:       opts,
		ShoppingItems: items,
		PriceHistory:  priceHistory,
		Decision:      dec,
	}, nil
}
