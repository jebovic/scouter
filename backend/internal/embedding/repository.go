package embedding

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository persists and queries embedding state on the options table.
type Repository interface {
	// SetEmbedding stores the embedding vector for the given option ID.
	SetEmbedding(ctx context.Context, id uuid.UUID, vec []float64) error
	// FindUnembeddedIDs returns up to limit option IDs with no embedding yet.
	FindUnembeddedIDs(ctx context.Context, limit int) ([]uuid.UUID, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a pgx-backed embedding repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// SetEmbedding updates the options.embedding column with the provided vector.
func (r *pgRepository) SetEmbedding(ctx context.Context, id uuid.UUID, vec []float64) error {
	const q = `UPDATE options SET embedding = $1::vector WHERE id = $2`
	tag, err := r.pool.Exec(ctx, q, FormatVector(vec), id)
	if err != nil {
		return fmt.Errorf("embedding: set: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("embedding: option %s not found", id)
	}
	return nil
}

// FindUnembeddedIDs returns option IDs that still have a NULL embedding column,
// ordered by creation time so the most recently added options are embedded first.
func (r *pgRepository) FindUnembeddedIDs(ctx context.Context, limit int) ([]uuid.UUID, error) {
	const q = `
		SELECT id FROM options
		WHERE embedding IS NULL
		ORDER BY created_at DESC
		LIMIT $1`
	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("embedding: find unembedded: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("embedding: scan id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
