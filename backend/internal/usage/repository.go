package usage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines persistence operations for llm_usage records.
type Repository interface {
	Insert(ctx context.Context, r Record) error
	Summarize(ctx context.Context, since time.Time) (Summary, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new usage Repository backed by PostgreSQL.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Insert(ctx context.Context, rec Record) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO llm_usage (provider, model, agent, mission_id, input_tokens, output_tokens, was_fallback)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		rec.Provider, rec.Model, rec.Agent, rec.MissionID,
		rec.InputTokens, rec.OutputTokens, rec.WasFallback,
	)
	return err
}

func (r *pgRepository) Summarize(ctx context.Context, since time.Time) (Summary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT provider, model,
		       COUNT(*)::int            AS calls,
		       SUM(input_tokens)::int   AS input_tokens,
		       SUM(output_tokens)::int  AS output_tokens
		FROM llm_usage
		WHERE created_at >= $1
		GROUP BY provider, model
		ORDER BY calls DESC`, since)
	if err != nil {
		return Summary{}, err
	}
	defer rows.Close()

	var s Summary
	for rows.Next() {
		var ps ProviderStat
		if err := rows.Scan(&ps.Provider, &ps.Model, &ps.Calls, &ps.InputTokens, &ps.OutputTokens); err != nil {
			return Summary{}, err
		}
		s.ByProvider = append(s.ByProvider, ps)
		s.TotalInput += ps.InputTokens
		s.TotalOutput += ps.OutputTokens
	}
	if err := rows.Err(); err != nil {
		return Summary{}, err
	}

	// Fallback calls count
	var fallbacks int
	err = r.pool.QueryRow(ctx, `
		SELECT COUNT(*)::int FROM llm_usage
		WHERE created_at >= $1 AND was_fallback = TRUE`, since).Scan(&fallbacks)
	if err != nil {
		return Summary{}, err
	}
	s.FallbackCalls = fallbacks

	// Ensure non-nil slice for JSON
	if s.ByProvider == nil {
		s.ByProvider = []ProviderStat{}
	}

	return s, nil
}

// UUIDPtr is a helper for optional mission_id FK — not exported; used by service.
func uuidPtr(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
