// backend/internal/researchjob/repository.go
package researchjob

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for research jobs.
type Repository interface {
	Create(ctx context.Context, missionID uuid.UUID, feedback string) (*ResearchJob, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, optionsCount int) (int64, error)
	GetByMission(ctx context.Context, missionID uuid.UUID, limit int) ([]ResearchJob, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ResearchJob, error)
	HasActiveJob(ctx context.Context, missionID uuid.UUID) (bool, error)
	FailStaleJobs(ctx context.Context) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed research job repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const jobCols = `id, mission_id, status, feedback, error, options_count, started_at, completed_at, created_at`

func (r *pgRepository) Create(ctx context.Context, missionID uuid.UUID, feedback string) (*ResearchJob, error) {
	var fb *string
	if feedback != "" {
		fb = &feedback
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO research_jobs (mission_id, feedback)
		VALUES ($1, $2)
		RETURNING `+jobCols,
		missionID, fb)
	return scanJob(row)
}

// UpdateStatus updates a job's status and sets timing/result fields.
// Returns (rowsAffected, error). rowsAffected == 0 means the job row no longer
// exists (mission was deleted via CASCADE).
func (r *pgRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, errMsg string, optionsCount int) (int64, error) {
	var tag interface{ RowsAffected() int64 }
	var err error

	switch status {
	case StatusRunning:
		now := time.Now()
		tag, err = r.pool.Exec(ctx, `
			UPDATE research_jobs
			SET status = $2, started_at = $3
			WHERE id = $1`, id, status, now)
	case StatusDone:
		now := time.Now()
		var cnt *int
		if optionsCount > 0 {
			cnt = &optionsCount
		}
		tag, err = r.pool.Exec(ctx, `
			UPDATE research_jobs
			SET status = $2, options_count = $3, completed_at = $4
			WHERE id = $1`, id, status, cnt, now)
	case StatusFailed:
		now := time.Now()
		var errPtr *string
		if errMsg != "" {
			errPtr = &errMsg
		}
		tag, err = r.pool.Exec(ctx, `
			UPDATE research_jobs
			SET status = $2, error = $3, completed_at = $4
			WHERE id = $1`, id, status, errPtr, now)
	default:
		return 0, fmt.Errorf("unknown status: %s", status)
	}

	if err != nil {
		return 0, fmt.Errorf("update research job status: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (r *pgRepository) GetByMission(ctx context.Context, missionID uuid.UUID, limit int) ([]ResearchJob, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.pool.Query(ctx, `
		SELECT `+jobCols+`
		FROM research_jobs
		WHERE mission_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, missionID, limit)
	if err != nil {
		return nil, fmt.Errorf("get research jobs by mission: %w", err)
	}
	defer rows.Close()

	var jobs []ResearchJob
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *j)
	}
	return jobs, rows.Err()
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*ResearchJob, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+jobCols+`
		FROM research_jobs
		WHERE id = $1`, id)
	j, err := scanJob(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return j, nil
}

func (r *pgRepository) HasActiveJob(ctx context.Context, missionID uuid.UUID) (bool, error) {
	var id uuid.UUID
	err := r.pool.QueryRow(ctx, `
		SELECT id FROM research_jobs
		WHERE mission_id = $1 AND status IN ('pending', 'running')
		LIMIT 1`, missionID).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("check active research job: %w", err)
	}
	return true, nil
}

func (r *pgRepository) FailStaleJobs(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE research_jobs
		SET status = 'failed', error = 'server restarted', completed_at = now()
		WHERE status IN ('pending', 'running')`)
	if err != nil {
		return fmt.Errorf("fail stale research jobs: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanJob(s scanner) (*ResearchJob, error) {
	var j ResearchJob
	err := s.Scan(
		&j.ID, &j.MissionID, &j.Status, &j.Feedback,
		&j.Error, &j.OptionsCount, &j.StartedAt, &j.CompletedAt, &j.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan research job: %w", err)
	}
	return &j, nil
}
