package decision

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for decisions.
type Repository interface {
	Create(ctx context.Context, d Decision) (*Decision, error)
	GetLatestByMission(ctx context.Context, missionID uuid.UUID) (*Decision, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed decision repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, d Decision) (*Decision, error) {
	scoresJSON, err := json.Marshal(d.Scores)
	if err != nil {
		return nil, fmt.Errorf("marshal scores: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO decisions (mission_id, scores, summary)
		VALUES ($1, $2, $3)
		RETURNING id, mission_id, scores, summary, created_at`,
		d.MissionID, scoresJSON, d.Summary)

	return scanDecision(row)
}

func (r *pgRepository) GetLatestByMission(ctx context.Context, missionID uuid.UUID) (*Decision, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, mission_id, scores, summary, created_at
		FROM decisions
		WHERE mission_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, missionID)

	d, err := scanDecision(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return d, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDecision(s rowScanner) (*Decision, error) {
	var d Decision
	var scoresRaw []byte

	err := s.Scan(&d.ID, &d.MissionID, &scoresRaw, &d.Summary, &d.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan decision: %w", err)
	}

	if len(scoresRaw) > 0 {
		if err := json.Unmarshal(scoresRaw, &d.Scores); err != nil {
			return nil, fmt.Errorf("unmarshal scores: %w", err)
		}
	}

	return &d, nil
}
