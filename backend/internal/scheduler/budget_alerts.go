package scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/notification"
	"log/slog"
)

// envelopeBudgetRepo is the subset of the envelope repository used for
// budget-alert checks. Defined here so tests can inject a stub.
type envelopeBudgetRepo interface {
	MonthlySpendAll(ctx context.Context) ([]envelope.EnvelopeSpend, error)
}

const (
	thresholdOver = 1.0
	thresholdWarn = 0.8
)

// checkBudgetAlerts queries every envelope with recorded spend this month and
// creates a notification when spend crosses the 80% or 100% threshold.
// Duplicate alerts within the current day are suppressed via ListFiltered.
func (o *Orchestrator) checkBudgetAlerts(ctx context.Context) {
	if o.envelopeRepo == nil {
		return
	}

	o.log.Info("budget alert check started")

	spends, err := o.envelopeRepo.MonthlySpendAll(ctx)
	if err != nil {
		o.log.Error("budget alert: fetch monthly spend", "err", err)
		return
	}

	for _, es := range spends {
		if ctx.Err() != nil {
			o.log.Warn("budget alert: context expired, stopping early")
			return
		}
		o.evaluateEnvelope(ctx, es)
	}

	o.log.Info("budget alert check complete", "envelopes_checked", len(spends))
}

// evaluateEnvelope decides whether a notification should be created for a
// single envelope and creates it when the threshold is crossed and no
// equivalent alert was sent today.
func (o *Orchestrator) evaluateEnvelope(ctx context.Context, es envelope.EnvelopeSpend) {
	budget := es.Envelope.MonthlyAmount
	if budget <= 0 {
		return
	}

	ratio := es.Spent / budget

	var title, body string
	switch {
	case ratio >= thresholdOver:
		title = fmt.Sprintf("Budget envelope '%s' dépassé!", es.Envelope.Name)
		// Embed envelope ID in body for dedup detection.
		body = fmt.Sprintf("Dépensé: %.2f€/%.2f€ [envelope:%s]", es.Spent, budget, es.Envelope.ID)
	case ratio >= thresholdWarn:
		title = fmt.Sprintf("Budget envelope '%s' à 80%%", es.Envelope.Name)
		// Embed envelope ID in body for dedup detection.
		body = fmt.Sprintf("Dépensé %.2f€ sur %.2f€ [envelope:%s]", es.Spent, budget, es.Envelope.ID)
	default:
		return // under 80% — nothing to do
	}

	if o.recentBudgetAlertExists(ctx, es.Envelope.ID) {
		o.log.Debug("budget alert: suppressed duplicate",
			"envelope_id", es.Envelope.ID,
			"envelope_name", es.Envelope.Name)
		return
	}

	_, err := o.notifRepo.Create(ctx, notification.CreateRequest{
		MissionID: es.MissionID,
		Type:      notification.TypeBudgetAlert,
		Title:     title,
		Body:      body,
	})
	if err != nil {
		o.log.Error("budget alert: create notification",
			"envelope_id", es.Envelope.ID,
			"err", err)
		return
	}

	o.recorder.RecordAlertTriggered()
	o.log.Info("budget alert created",
		"envelope_name", es.Envelope.Name,
		slog.Float64("ratio", ratio))
}

// recentBudgetAlertExists returns true when a budget_alert notification whose
// body contains "[envelope:<id>]" was already created today.
// This avoids creating duplicate alerts within the same calendar day.
func (o *Orchestrator) recentBudgetAlertExists(ctx context.Context, envelopeID uuid.UUID) bool {
	t := notification.TypeBudgetAlert
	recents, err := o.notifRepo.ListFiltered(ctx, notification.ListFilter{
		Type:  &t,
		Limit: 200,
	})
	if err != nil {
		// Fail open: if we cannot check, allow the alert to proceed.
		o.log.Warn("budget alert: list filtered failed, proceeding", "err", err)
		return false
	}

	marker := fmt.Sprintf("[envelope:%s]", envelopeID)
	for _, n := range recents {
		if strings.Contains(n.Body, marker) {
			return true
		}
	}
	return false
}
