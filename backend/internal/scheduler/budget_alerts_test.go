package scheduler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/envelope"
	"github.com/jibei/scouter/internal/notification"
)

// ---- mock envelope budget repo ----

type mockEnvelopeBudgetRepo struct {
	spends []envelope.EnvelopeSpend
	err    error
}

func (m *mockEnvelopeBudgetRepo) MonthlySpendAll(_ context.Context) ([]envelope.EnvelopeSpend, error) {
	return m.spends, m.err
}

// ---- helpers ----

func envelopeSpend(name string, monthly, spent float64) envelope.EnvelopeSpend {
	return envelope.EnvelopeSpend{
		Envelope: envelope.Envelope{
			ID:            uuid.New(),
			Name:          name,
			MonthlyAmount: monthly,
			Currency:      "EUR",
			CreatedAt:     time.Now(),
		},
		Spent:     spent,
		MissionID: uuid.New(),
	}
}

func orchestratorWithEnvelope(
	envRepo envelopeBudgetRepo,
	notifRepo notification.Repository,
) *Orchestrator {
	o := newTestOrchestrator(
		&mockMissionRepo{},
		&mockOptionRepo{},
		&mockShoppingRepo{},
		notifRepo,
		&mockPricingRunner{},
		10,
	)
	o.SetEnvelopeRepo(envRepo)
	return o
}

// ---- tests ----

func TestCheckBudgetAlerts_NoEnvelopeRepo(t *testing.T) {
	// When envelopeRepo is nil, checkBudgetAlerts must return without panicking.
	notifRepo := &mockOrchestratorNotifRepo{}
	o := newTestOrchestrator(
		&mockMissionRepo{},
		&mockOptionRepo{},
		&mockShoppingRepo{},
		notifRepo,
		&mockPricingRunner{},
		10,
	)
	// envelopeRepo intentionally left nil
	o.checkBudgetAlerts(context.Background())
	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications when envelopeRepo is nil, got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_RepoError(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	envRepo := &mockEnvelopeBudgetRepo{err: errors.New("db error")}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications on repo error, got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_NothingOver80Pct(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	envRepo := &mockEnvelopeBudgetRepo{
		spends: []envelope.EnvelopeSpend{
			envelopeSpend("Groceries", 300, 200), // 66% — under threshold
		},
	}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications under 80%%, got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_Warn80Pct(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	es := envelopeSpend("Groceries", 300, 250) // 83.3% — warn threshold
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifRepo.created))
	}
	n := notifRepo.created[0]
	if n.Type != notification.TypeBudgetAlert {
		t.Errorf("type = %q; want %q", n.Type, notification.TypeBudgetAlert)
	}
	if !strings.Contains(n.Title, "80%") {
		t.Errorf("title %q should mention 80%%", n.Title)
	}
	if !strings.Contains(n.Title, "Groceries") {
		t.Errorf("title %q should mention envelope name", n.Title)
	}
	marker := fmt.Sprintf("[envelope:%s]", es.Envelope.ID)
	if !strings.Contains(n.Body, marker) {
		t.Errorf("body %q should contain envelope ID marker %q for dedup", n.Body, marker)
	}
	if n.MissionID != es.MissionID {
		t.Errorf("mission_id = %v; want %v", n.MissionID, es.MissionID)
	}
}

func TestCheckBudgetAlerts_Over100Pct(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	es := envelopeSpend("Rent", 1000, 1100) // 110% — exceeded threshold
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifRepo.created))
	}
	n := notifRepo.created[0]
	if n.Type != notification.TypeBudgetAlert {
		t.Errorf("type = %q; want %q", n.Type, notification.TypeBudgetAlert)
	}
	if !strings.Contains(n.Title, "dépassé") {
		t.Errorf("title %q should indicate budget exceeded", n.Title)
	}
	if !strings.Contains(n.Title, "Rent") {
		t.Errorf("title %q should mention envelope name", n.Title)
	}
}

func TestCheckBudgetAlerts_Exactly80Pct(t *testing.T) {
	// 80% exactly should trigger the warn notification.
	notifRepo := &mockOrchestratorNotifRepo{}
	es := envelopeSpend("Transport", 200, 160) // exactly 80%
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification at exactly 80%%, got %d", len(notifRepo.created))
	}
	if !strings.Contains(notifRepo.created[0].Title, "80%") {
		t.Errorf("title %q should mention 80%%", notifRepo.created[0].Title)
	}
}

func TestCheckBudgetAlerts_Exactly100Pct(t *testing.T) {
	// 100% exactly should trigger the exceeded notification.
	notifRepo := &mockOrchestratorNotifRepo{}
	es := envelopeSpend("Leisure", 500, 500) // exactly 100%
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification at exactly 100%%, got %d", len(notifRepo.created))
	}
	if !strings.Contains(notifRepo.created[0].Title, "dépassé") {
		t.Errorf("title %q should indicate budget exceeded at 100%%", notifRepo.created[0].Title)
	}
}

func TestCheckBudgetAlerts_MultipleEnvelopes(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	envRepo := &mockEnvelopeBudgetRepo{
		spends: []envelope.EnvelopeSpend{
			envelopeSpend("Groceries", 300, 50),   // 16% — silent
			envelopeSpend("Rent", 1000, 900),       // 90% — warn
			envelopeSpend("Travel", 200, 210),      // 105% — exceeded
		},
	}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 2 {
		t.Fatalf("expected 2 notifications (one warn + one exceeded), got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_DedupSuppressesSecondAlert(t *testing.T) {
	es := envelopeSpend("Rent", 1000, 900) // 90% — warn

	// Pre-populate the notif repo with a prior alert for this envelope.
	marker := fmt.Sprintf("[envelope:%s]", es.Envelope.ID)
	existingNotif := notification.Notification{
		ID:        uuid.New(),
		MissionID: es.MissionID,
		Type:      notification.TypeBudgetAlert,
		Title:     fmt.Sprintf("Budget envelope '%s' à 80%%", es.Envelope.Name),
		Body:      fmt.Sprintf("Dépensé 900.00€ sur 1000.00€ %s", marker),
		CreatedAt: time.Now(),
	}

	notifRepo := &mockOrchestratorNotifRepo{}
	// Wire the existing notification so ListFiltered returns it.
	notifRepo.existing = []notification.Notification{existingNotif}

	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 new notifications (duplicate suppressed), got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_ZeroBudget_Skipped(t *testing.T) {
	// Envelopes with monthly_amount = 0 must not cause a division-by-zero or alert.
	notifRepo := &mockOrchestratorNotifRepo{}
	es := envelopeSpend("Ghost", 0, 50)
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications for zero-budget envelope, got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_ListFilteredError_FailOpen(t *testing.T) {
	// When ListFiltered returns an error, the alert should still be created (fail open).
	notifRepo := &mockOrchestratorNotifRepo{listFilteredErr: errors.New("db error")}
	es := envelopeSpend("Food", 400, 360) // 90%
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	o.checkBudgetAlerts(context.Background())

	if len(notifRepo.created) != 1 {
		t.Fatalf("expected 1 notification on ListFiltered error (fail open), got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_CreateError_LogsAndContinues(t *testing.T) {
	// When Create returns an error, the function should log and not panic.
	notifRepo := &mockOrchestratorNotifRepo{createErr: errors.New("insert failed")}
	es := envelopeSpend("Bills", 300, 270) // 90%
	envRepo := &mockEnvelopeBudgetRepo{spends: []envelope.EnvelopeSpend{es}}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	// Should not panic.
	o.checkBudgetAlerts(context.Background())

	// No notification recorded (Create failed).
	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 created notifications on Create error, got %d", len(notifRepo.created))
	}
}

func TestCheckBudgetAlerts_ContextCancelled_StopsEarly(t *testing.T) {
	notifRepo := &mockOrchestratorNotifRepo{}
	// Three envelopes all above threshold — with cancelled context only zero should be processed.
	envRepo := &mockEnvelopeBudgetRepo{
		spends: []envelope.EnvelopeSpend{
			envelopeSpend("A", 100, 90),
			envelopeSpend("B", 100, 90),
			envelopeSpend("C", 100, 90),
		},
	}
	o := orchestratorWithEnvelope(envRepo, notifRepo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before calling

	o.checkBudgetAlerts(ctx)

	// With cancelled context the loop exits on first iteration check.
	if len(notifRepo.created) != 0 {
		t.Errorf("expected 0 notifications with cancelled context, got %d", len(notifRepo.created))
	}
}
