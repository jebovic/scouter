package collaborator

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const collabCols = `id, mission_id, invite_token, email, display_name, role, joined_at, created_at, updated_at`

// Repository handles persistence of collaborators.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed collaborator repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new collaborator invite and returns the full record.
func (r *Repository) Create(ctx context.Context, missionID, email string, role Role) (*Collaborator, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO collaborators (mission_id, email, role)
		VALUES ($1, $2, $3)
		RETURNING `+collabCols,
		missionID, email, string(role))
	c, err := scanCollaborator(row)
	if err != nil {
		return nil, fmt.Errorf("collaborator: create: %w", err)
	}
	return c, nil
}

// GetByInviteToken fetches a collaborator by its invite token.
// Returns pgx.ErrNoRows when no match is found.
func (r *Repository) GetByInviteToken(ctx context.Context, token string) (*Collaborator, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+collabCols+`
		FROM collaborators
		WHERE invite_token = $1`, token)
	c, err := scanCollaborator(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("collaborator: get by token: %w", err)
	}
	return c, nil
}

// MarkJoined sets joined_at and display_name for the collaborator identified by
// token. The operation is idempotent — calling it multiple times is safe.
func (r *Repository) MarkJoined(ctx context.Context, token, displayName string) (*Collaborator, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE collaborators
		SET joined_at = COALESCE(joined_at, NOW()),
		    display_name = $2,
		    updated_at = NOW()
		WHERE invite_token = $1
		RETURNING `+collabCols,
		token, displayName)
	c, err := scanCollaborator(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("collaborator: mark joined: %w", err)
	}
	return c, nil
}

// ListByMission returns all collaborators for the given mission ordered by creation time.
func (r *Repository) ListByMission(ctx context.Context, missionID string) ([]Collaborator, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+collabCols+`
		FROM collaborators
		WHERE mission_id = $1
		ORDER BY created_at ASC`, missionID)
	if err != nil {
		return nil, fmt.Errorf("collaborator: list by mission: %w", err)
	}
	defer rows.Close()

	var out []Collaborator
	for rows.Next() {
		c, err := scanCollaborator(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

// GetByID fetches a single collaborator by its primary key.
func (r *Repository) GetByID(ctx context.Context, id string) (*Collaborator, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+collabCols+`
		FROM collaborators
		WHERE id = $1`, id)
	c, err := scanCollaborator(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("collaborator: get by id: %w", err)
	}
	return c, nil
}

// Delete removes a collaborator record by ID.
func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM collaborators WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("collaborator: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCollaborator(s scanner) (*Collaborator, error) {
	var c Collaborator
	err := s.Scan(
		&c.ID,
		&c.MissionID,
		&c.InviteToken,
		&c.Email,
		&c.DisplayName,
		&c.Role,
		&c.JoinedAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
