package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/option"
)

// Repository performs vector similarity queries against the options table.
type Repository interface {
	// Search returns the top-limit options closest to vec using cosine distance.
	Search(ctx context.Context, vec []float64, limit int) ([]Result, error)
	// Similar returns options similar to the given optionID, excluding itself.
	Similar(ctx context.Context, optionID uuid.UUID, limit int) ([]Result, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a pgx-backed search repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// Search performs an ANN search using the IVFFlat index on the embedding column.
// Options without an embedding are excluded automatically (NULL <=> vec is NULL).
func (r *pgRepository) Search(ctx context.Context, vec []float64, limit int) ([]Result, error) {
	const q = `
		SELECT o.id, o.mission_id, m.name, m.slug,
		       o.name, o.category, o.badge, o.price_range, o.notes,
		       1 - (o.embedding <=> $1::vector) AS similarity
		FROM options o
		JOIN missions m ON m.id = o.mission_id
		WHERE o.embedding IS NOT NULL
		ORDER BY o.embedding <=> $1::vector
		LIMIT $2`

	rows, err := r.pool.Query(ctx, q, embedding.FormatVector(vec), limit)
	if err != nil {
		return nil, fmt.Errorf("search: query: %w", err)
	}
	defer rows.Close()
	return scanResults(rows)
}

// Similar finds options whose embeddings are closest to optionID's embedding.
func (r *pgRepository) Similar(ctx context.Context, optionID uuid.UUID, limit int) ([]Result, error) {
	const q = `
		WITH target AS (
			SELECT embedding FROM options WHERE id = $1 AND embedding IS NOT NULL
		)
		SELECT o.id, o.mission_id, m.name, m.slug,
		       o.name, o.category, o.badge, o.price_range, o.notes,
		       1 - (o.embedding <=> t.embedding) AS similarity
		FROM options o
		JOIN missions m ON m.id = o.mission_id
		CROSS JOIN target t
		WHERE o.embedding IS NOT NULL
		  AND o.id <> $1
		ORDER BY o.embedding <=> t.embedding
		LIMIT $2`

	rows, err := r.pool.Query(ctx, q, optionID, limit)
	if err != nil {
		return nil, fmt.Errorf("search similar: query: %w", err)
	}
	defer rows.Close()
	return scanResults(rows)
}

type scannable interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

// scanResults reads Result rows from pgx.Rows.
func scanResults(rows scannable) ([]Result, error) {
	var out []Result
	for rows.Next() {
		var res Result
		var priceJSON []byte
		if err := rows.Scan(
			&res.ID, &res.MissionID, &res.MissionName, &res.MissionSlug,
			&res.Name, &res.Category, &res.Badge, &priceJSON, &res.Notes,
			&res.Similarity,
		); err != nil {
			return nil, fmt.Errorf("search: scan: %w", err)
		}
		if len(priceJSON) > 0 {
			var pr option.PriceRange
			if err := json.Unmarshal(priceJSON, &pr); err != nil {
				log.Printf("search: unmarshal price_range for option %s: %v", res.ID, err)
			} else {
				res.PriceRange = &pr
			}
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

