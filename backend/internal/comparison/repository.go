package comparison

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles persistence for comparison weights.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new comparison Repository backed by pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// SetWeight upserts a criterion weight for a mission.
func (r *Repository) SetWeight(ctx context.Context, missionID string, req SetWeightRequest) (Weight, error) {
	if err := validateCriteria(req.Criteria); err != nil {
		return Weight{}, err
	}
	if err := validateWeight(req.Weight); err != nil {
		return Weight{}, err
	}

	var w Weight
	err := r.pool.QueryRow(ctx, `
		INSERT INTO comparison_weights (mission_id, criteria, weight)
		VALUES ($1, $2, $3)
		ON CONFLICT (mission_id, criteria) DO UPDATE
			SET weight = $3, updated_at = NOW()
		RETURNING id, mission_id, criteria, weight, created_at, updated_at
	`, missionID, req.Criteria, req.Weight).Scan(
		&w.ID, &w.MissionID, &w.Criteria, &w.Weight, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return Weight{}, fmt.Errorf("set weight: %w", err)
	}
	return w, nil
}

// ListWeights returns all weights for a mission ordered by criteria name.
func (r *Repository) ListWeights(ctx context.Context, missionID string) ([]Weight, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, mission_id, criteria, weight, created_at, updated_at
		FROM comparison_weights
		WHERE mission_id = $1
		ORDER BY criteria ASC
	`, missionID)
	if err != nil {
		return nil, fmt.Errorf("list weights: %w", err)
	}
	defer rows.Close()

	weights := []Weight{}
	for rows.Next() {
		var w Weight
		if err := rows.Scan(&w.ID, &w.MissionID, &w.Criteria, &w.Weight, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan weight: %w", err)
		}
		weights = append(weights, w)
	}
	return weights, rows.Err()
}

// DeleteWeight removes a criterion weight for a mission.
// Returns pgx.ErrNoRows if the weight did not exist.
func (r *Repository) DeleteWeight(ctx context.Context, missionID, criteria string) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM comparison_weights WHERE mission_id = $1 AND criteria = $2
	`, missionID, criteria)
	if err != nil {
		return fmt.Errorf("delete weight: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// ---------------------------------------------------------------------------
// Validation helpers (exported for tests in the same package)
// ---------------------------------------------------------------------------

func validateWeight(w int) error {
	if w < 0 || w > 100 {
		return errors.New("weight must be between 0 and 100")
	}
	return nil
}

func validateCriteria(c string) error {
	if strings.TrimSpace(c) == "" {
		return errors.New("criteria must not be empty")
	}
	return nil
}
