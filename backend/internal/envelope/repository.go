package envelope

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const envelopeCols = `id, name, monthly_amount, currency, created_at`

// Repository handles all envelope persistence via pgxpool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a Repository backed by the given pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scanEnvelope(s interface{ Scan(...any) error }) (Envelope, error) {
	var e Envelope
	err := s.Scan(&e.ID, &e.Name, &e.MonthlyAmount, &e.Currency, &e.CreatedAt)
	return e, err
}

// FindAll returns every envelope ordered by name.
func (r *Repository) FindAll(ctx context.Context) ([]Envelope, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+envelopeCols+` FROM envelopes ORDER BY name ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Envelope{}
	for rows.Next() {
		e, err := scanEnvelope(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// Create inserts a new envelope and returns the persisted record.
func (r *Repository) Create(ctx context.Context, e Envelope) (Envelope, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO envelopes (name, monthly_amount, currency)
		 VALUES ($1, $2, $3)
		 RETURNING `+envelopeCols,
		e.Name, e.MonthlyAmount, e.Currency,
	)
	return scanEnvelope(row)
}

// Update applies name / monthly_amount / currency changes to an existing envelope.
func (r *Repository) Update(ctx context.Context, e Envelope) (Envelope, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE envelopes
		 SET name = $1, monthly_amount = $2, currency = $3
		 WHERE id = $4
		 RETURNING `+envelopeCols,
		e.Name, e.MonthlyAmount, e.Currency, e.ID,
	)
	updated, err := scanEnvelope(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Envelope{}, pgx.ErrNoRows
	}
	return updated, err
}

// GetSummary returns the envelope and aggregated mission stats for the given envelope ID.
// Returns nil if the envelope does not exist.
func (r *Repository) GetSummary(ctx context.Context, id uuid.UUID) (*EnvelopeSummary, error) {
	const q = `
	WITH mission_stats AS (
		SELECT
			COUNT(*)                                        AS mission_count,
			COALESCE(SUM(m.budget), 0)                      AS total_budget,
			COALESCE(SUM(pr.final_price), 0)                AS total_spent
		FROM missions m
		LEFT JOIN purchase_records pr ON pr.mission_id = m.id
		WHERE m.envelope_id = $1
	)
	SELECT
		e.id, e.name, e.monthly_amount, e.currency, e.created_at,
		ms.mission_count, ms.total_budget, ms.total_spent
	FROM envelopes e, mission_stats ms
	WHERE e.id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	var s EnvelopeSummary
	err := row.Scan(
		&s.Envelope.ID, &s.Envelope.Name, &s.Envelope.MonthlyAmount,
		&s.Envelope.Currency, &s.Envelope.CreatedAt,
		&s.MissionCount, &s.TotalBudget, &s.TotalSpent,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get envelope summary: %w", err)
	}
	return &s, nil
}

// Delete removes an envelope by UUID; returns pgx.ErrNoRows if it did not exist.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM envelopes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
