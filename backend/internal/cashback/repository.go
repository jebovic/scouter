package cashback

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrNotFound is returned when an entry does not exist.
var ErrNotFound = errors.New("cashback entry not found")

// Repository is an in-memory store for cashback entries.
type Repository struct {
	mu      sync.RWMutex
	entries map[uuid.UUID]*Entry
}

// NewRepository creates a new in-memory cashback repository.
func NewRepository() *Repository {
	return &Repository{
		entries: make(map[uuid.UUID]*Entry),
	}
}

// Create stores a new entry and returns it.
func (r *Repository) Create(userRef string, req CreateRequest) (*Entry, error) {
	status := req.Status
	if status == "" {
		status = StatusPending
	}

	now := time.Now().UTC()
	e := &Entry{
		ID:          uuid.New(),
		UserRef:     userRef,
		ItemName:    req.ItemName,
		Merchant:    req.Merchant,
		Amount:      req.Amount,
		PercentRate: req.PercentRate,
		Status:      status,
		Notes:       req.Notes,
		CreatedAt:   now,
	}

	r.mu.Lock()
	r.entries[e.ID] = e
	r.mu.Unlock()

	return e, nil
}

// List returns all entries for a given userRef, ordered by creation time descending.
func (r *Repository) List(userRef string) ([]*Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Entry
	for _, e := range r.entries {
		if e.UserRef == userRef {
			copied := *e
			result = append(result, &copied)
		}
	}

	// Sort by createdAt descending (newest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].CreatedAt.After(result[i].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// GetByID returns a single entry by ID.
func (r *Repository) GetByID(id uuid.UUID) (*Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	e, ok := r.entries[id]
	if !ok {
		return nil, ErrNotFound
	}
	copied := *e
	return &copied, nil
}

// Update applies status, notes, and paidAt changes to an existing entry.
func (r *Repository) Update(id uuid.UUID, req UpdateRequest) (*Entry, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	e, ok := r.entries[id]
	if !ok {
		return nil, ErrNotFound
	}

	if req.Status != "" {
		e.Status = req.Status
	}
	if req.Notes != "" {
		e.Notes = req.Notes
	}
	if req.PaidAt != nil {
		e.PaidAt = req.PaidAt
	}

	copied := *e
	return &copied, nil
}

// Delete removes an entry by ID.
func (r *Repository) Delete(id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.entries[id]; !ok {
		return ErrNotFound
	}
	delete(r.entries, id)
	return nil
}

// Summary returns aggregated totals for a given userRef.
func (r *Repository) Summary(userRef string) (*Summary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s := &Summary{}
	for _, e := range r.entries {
		if e.UserRef != userRef {
			continue
		}
		s.EntryCount++
		switch e.Status {
		case StatusPending:
			s.TotalPending += e.Amount
		case StatusConfirmed:
			s.TotalConfirmed += e.Amount
		case StatusPaid:
			s.TotalPaid += e.Amount
		}
	}
	return s, nil
}
