package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

// Scheduler wraps robfig/cron and runs the Orchestrator on a cron schedule.
type Scheduler struct {
	cron         *cron.Cron
	orchestrator *Orchestrator
	log          *slog.Logger
}

// New creates a new Scheduler. Call Start to begin scheduling.
func New(cronExpr string, o *Orchestrator, log *slog.Logger) (*Scheduler, error) {
	c := cron.New(cron.WithLocation(time.UTC))
	s := &Scheduler{cron: c, orchestrator: o, log: log}

	_, err := c.AddFunc(cronExpr, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
		defer cancel()
		o.RunPriceChecks(ctx)
	})
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Start begins the cron scheduler.
func (s *Scheduler) Start() {
	s.log.Info("price check scheduler started")
	s.cron.Start()
}

// Stop drains the scheduler, waiting for any running job to finish (up to 60s).
func (s *Scheduler) Stop() {
	s.log.Info("price check scheduler stopping")
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
		s.log.Info("price check scheduler stopped cleanly")
	case <-time.After(60 * time.Second):
		s.log.Warn("price check scheduler stop timed out")
	}
}
