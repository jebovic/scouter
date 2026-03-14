// Package scheduler implements the in-process price-check cron scheduler.
// It lists active missions, runs the PricingAgent for each, and creates
// notifications when target prices are hit or significant price drops occur.
package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jibei/scouter/internal/dealintel"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/notification"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/pricing"
	"github.com/jibei/scouter/internal/shopping"
)

// Orchestrator runs periodic price checks across active missions.
type Orchestrator struct {
	missionRepo  mission.Repository
	optionRepo   option.Repository
	shoppingRepo shopping.Repository
	notifRepo    notification.Repository
	pricingAgent *pricing.Agent
	maxMissions  int
	log          *slog.Logger
}

// NewOrchestrator creates a new price-check orchestrator.
func NewOrchestrator(
	missionRepo mission.Repository,
	optionRepo option.Repository,
	shoppingRepo shopping.Repository,
	notifRepo notification.Repository,
	pricingAgent *pricing.Agent,
	maxMissions int,
	log *slog.Logger,
) *Orchestrator {
	return &Orchestrator{
		missionRepo:  missionRepo,
		optionRepo:   optionRepo,
		shoppingRepo: shoppingRepo,
		notifRepo:    notifRepo,
		pricingAgent: pricingAgent,
		maxMissions:  maxMissions,
		log:          log,
	}
}

// RunPriceChecks lists active missions and runs a price check for each.
func (o *Orchestrator) RunPriceChecks(ctx context.Context) {
	o.log.Info("price check run started")
	start := time.Now()

	missions, err := o.missionRepo.ListActive(ctx)
	if err != nil {
		o.log.Error("price check: list active missions", "err", err)
		return
	}

	limit := o.maxMissions
	if len(missions) < limit {
		limit = len(missions)
	}

	checked := 0
	for _, m := range missions[:limit] {
		if ctx.Err() != nil {
			o.log.Warn("price check: context expired, missions skipped",
				"missions_checked", checked, "missions_total", limit)
			break
		}
		if err := o.checkMission(ctx, m); err != nil {
			o.log.Error("price check: mission failed", "mission_id", m.ID, "slug", m.Slug, "err", err)
		}
		checked++
	}

	o.log.Info("price check run complete",
		"missions_checked", checked, "missions_total", limit, "duration", time.Since(start))
}

func (o *Orchestrator) checkMission(ctx context.Context, m mission.Mission) error {
	opts, err := o.optionRepo.ListByMission(ctx, m.ID)
	if err != nil {
		return fmt.Errorf("list options: %w", err)
	}
	if len(opts) == 0 {
		return nil
	}

	items, err := o.pricingAgent.Run(ctx, m, opts, nil)
	if err != nil {
		return fmt.Errorf("pricing agent: %w", err)
	}

	for _, item := range items {
		if err := o.maybeNotify(ctx, m, item); err != nil {
			o.log.Warn("price check: notify failed", "item_id", item.ID, "err", err)
		}
	}
	return nil
}

func (o *Orchestrator) maybeNotify(ctx context.Context, m mission.Mission, item shopping.Item) error {
	snapshots, err := o.shoppingRepo.ListPriceHistory(ctx, item.ID)
	if err != nil {
		return err
	}

	points := make([]dealintel.PricePoint, len(snapshots))
	for i, s := range snapshots {
		points[i] = dealintel.PricePoint{Price: s.Price, RecordedAt: s.RecordedAt}
	}

	score := dealintel.ComputeDealScore(item.Price, item.TargetPrice, points)
	if score == nil {
		return nil
	}

	itemID := item.ID
	// PctBelowTarget > 0 means current price is below target — target hit.
	if item.TargetPrice != nil && score.PctBelowTarget > 0 {
		_, err = o.notifRepo.Create(ctx, notification.CreateRequest{
			MissionID: m.ID,
			ItemID:    &itemID,
			Type:      "target_hit",
			Title:     fmt.Sprintf("Target price hit: %s", item.Name),
			Body:      fmt.Sprintf("Current price %.2f has reached your target of %.2f.", item.Price, *item.TargetPrice),
		})
		if err != nil {
			return fmt.Errorf("create target_hit notification: %w", err)
		}
		return nil
	}

	// PctBelowAvg > 0 and dropping trend means a meaningful price drop.
	if score.Trend == dealintel.TrendDropping && score.PctBelowAvg > 0 {
		_, err = o.notifRepo.Create(ctx, notification.CreateRequest{
			MissionID: m.ID,
			ItemID:    &itemID,
			Type:      "price_drop",
			Title:     fmt.Sprintf("Price drop: %s", item.Name),
			Body:      fmt.Sprintf("Price dropped to %.2f (%.1f%% below average).", item.Price, score.PctBelowAvg),
		})
		if err != nil {
			return fmt.Errorf("create price_drop notification: %w", err)
		}
	}

	return nil
}
