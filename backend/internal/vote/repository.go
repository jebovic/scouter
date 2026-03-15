package vote

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles persistence of option votes.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed vote repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Record upserts a vote for an option by a collaborator.
// Existing votes are updated to the new value.
func (r *Repository) Record(ctx context.Context, optionID, collaboratorID string, v int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO option_votes (option_id, collaborator_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT (option_id, collaborator_id)
		DO UPDATE SET vote = EXCLUDED.vote, updated_at = NOW()`,
		optionID, collaboratorID, v)
	if err != nil {
		return fmt.Errorf("vote: record: %w", err)
	}
	return nil
}

// GetAggregated returns the aggregated vote counts and per-collaborator map for
// the given option.
func (r *Repository) GetAggregated(ctx context.Context, optionID string) (*AggregatedVotes, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT collaborator_id, vote
		FROM option_votes
		WHERE option_id = $1`, optionID)
	if err != nil {
		return nil, fmt.Errorf("vote: get aggregated: %w", err)
	}
	defer rows.Close()

	result := &AggregatedVotes{
		OptionID:       optionID,
		ByCollaborator: make(map[string]int),
	}

	for rows.Next() {
		var collabID string
		var voteVal int
		if err := rows.Scan(&collabID, &voteVal); err != nil {
			return nil, fmt.Errorf("vote: scan row: %w", err)
		}
		result.ByCollaborator[collabID] = voteVal
		switch voteVal {
		case 1:
			result.Upvotes++
		case -1:
			result.Downvotes++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("vote: iterate rows: %w", err)
	}

	result.Score = result.Upvotes - result.Downvotes
	return result, nil
}
