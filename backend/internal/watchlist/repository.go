package watchlist

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	mu      sync.RWMutex
	entries map[uuid.UUID]*WatchEntry
}

func NewRepository() *Repository {
	return &Repository{entries: make(map[uuid.UUID]*WatchEntry)}
}

func (r *Repository) List(_ context.Context) []*WatchEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*WatchEntry, 0, len(r.entries))
	for _, e := range r.entries {
		out = append(out, e)
	}
	return out
}

func (r *Repository) Create(_ context.Context, req CreateRequest) (*WatchEntry, error) {
	itemID, err := uuid.Parse(req.ItemID)
	if err != nil {
		return nil, ErrInvalidItemID
	}
	e := &WatchEntry{
		ID:            uuid.New(),
		ItemID:        itemID,
		ItemName:      req.ItemName,
		BaselinePrice: req.BaselinePrice,
		ThresholdPct:  req.ThresholdPct,
		CurrentPrice:  req.BaselinePrice,
		CreatedAt:     time.Now(),
	}
	r.mu.Lock()
	r.entries[e.ID] = e
	r.mu.Unlock()
	return e, nil
}

func (r *Repository) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.entries[id]; !ok {
		return ErrNotFound
	}
	delete(r.entries, id)
	return nil
}

func (r *Repository) UpdatePrice(_ context.Context, itemID uuid.UUID, price float64) *CheckResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, e := range r.entries {
		if e.ItemID != itemID {
			continue
		}
		e.CurrentPrice = price
		dropPct := (e.BaselinePrice - price) / e.BaselinePrice * 100
		if dropPct >= e.ThresholdPct && e.TriggeredAt == nil {
			now := time.Now()
			e.TriggeredAt = &now
			return &CheckResult{
				Triggered: true,
				DropPct:   dropPct,
				Message:   fmt.Sprintf("Prix baissé de %.1f%% sur %s", dropPct, e.ItemName),
			}
		}
	}
	return &CheckResult{Triggered: false}
}
