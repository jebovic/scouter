package agentrun

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for agent runs.
type Repository interface {
	Create(ctx context.Context, run Run) (*Run, error)
	ListByMission(ctx context.Context, missionID uuid.UUID, agentType *AgentType) ([]Run, error)
	GetLatestByMission(ctx context.Context, missionID uuid.UUID, agentType AgentType) (*Run, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed agent run repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Create(ctx context.Context, run Run) (*Run, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO agent_runs (mission_id, agent_type, input_snapshot, result_snapshot, diff, feedback)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, mission_id, agent_type, input_snapshot, result_snapshot, diff, feedback, created_at`,
		run.MissionID, string(run.AgentType),
		[]byte(run.InputSnapshot), []byte(run.ResultSnapshot), []byte(run.Diff),
		run.Feedback)

	return scanRun(row)
}

func (r *pgRepository) ListByMission(ctx context.Context, missionID uuid.UUID, agentType *AgentType) ([]Run, error) {
	var rows pgx.Rows
	var err error

	if agentType != nil {
		rows, err = r.pool.Query(ctx, `
			SELECT id, mission_id, agent_type, input_snapshot, result_snapshot, diff, feedback, created_at
			FROM agent_runs
			WHERE mission_id = $1 AND agent_type = $2
			ORDER BY created_at DESC`, missionID, string(*agentType))
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, mission_id, agent_type, input_snapshot, result_snapshot, diff, feedback, created_at
			FROM agent_runs
			WHERE mission_id = $1
			ORDER BY created_at DESC`, missionID)
	}
	if err != nil {
		return nil, fmt.Errorf("list agent runs: %w", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, fmt.Errorf("list agent runs: scan: %w", err)
		}
		runs = append(runs, *run)
	}
	if runs == nil {
		runs = []Run{}
	}
	return runs, rows.Err()
}

func (r *pgRepository) GetLatestByMission(ctx context.Context, missionID uuid.UUID, agentType AgentType) (*Run, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, mission_id, agent_type, input_snapshot, result_snapshot, diff, feedback, created_at
		FROM agent_runs
		WHERE mission_id = $1 AND agent_type = $2
		ORDER BY created_at DESC
		LIMIT 1`, missionID, string(agentType))

	run, err := scanRun(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return run, err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRun(s scanner) (*Run, error) {
	var run Run
	var agentType string
	var inputRaw, resultRaw, diffRaw []byte
	var feedback *string

	err := s.Scan(
		&run.ID, &run.MissionID, &agentType,
		&inputRaw, &resultRaw, &diffRaw,
		&feedback, &run.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan agent run: %w", err)
	}

	run.AgentType = AgentType(agentType)
	if feedback != nil {
		run.Feedback = *feedback
	}
	if len(inputRaw) > 0 {
		run.InputSnapshot = inputRaw
	} else {
		run.InputSnapshot = []byte("{}")
	}
	if len(resultRaw) > 0 {
		run.ResultSnapshot = resultRaw
	} else {
		run.ResultSnapshot = []byte("{}")
	}
	if len(diffRaw) > 0 {
		run.Diff = diffRaw
	} else {
		run.Diff = []byte("{}")
	}

	return &run, nil
}
