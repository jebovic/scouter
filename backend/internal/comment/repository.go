package comment

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for mission comments.
type Repository interface {
	List(ctx context.Context, missionID string, optionID *string) ([]Comment, error)
	Create(ctx context.Context, missionID string, req CreateCommentRequest) (Comment, error)
	Delete(ctx context.Context, id string) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed comment repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const commentCols = `id, mission_id, option_id, author, body, created_at, updated_at`

func (r *pgRepository) List(ctx context.Context, missionID string, optionID *string) ([]Comment, error) {
	var (
		rows interface {
			Next() bool
			Scan(...any) error
			Close()
			Err() error
		}
		err error
	)
	if optionID != nil {
		rows, err = r.pool.Query(ctx, `
			SELECT `+commentCols+`
			FROM mission_comments
			WHERE mission_id = $1 AND option_id = $2
			ORDER BY created_at DESC`,
			missionID, *optionID)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT `+commentCols+`
			FROM mission_comments
			WHERE mission_id = $1
			ORDER BY created_at DESC`,
			missionID)
	}
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		c, scanErr := scanComment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		comments = append(comments, *c)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list comments rows: %w", rows.Err())
	}
	return comments, nil
}

func (r *pgRepository) Create(ctx context.Context, missionID string, req CreateCommentRequest) (Comment, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO mission_comments (mission_id, option_id, author, body)
		VALUES ($1, $2, $3, $4)
		RETURNING `+commentCols,
		missionID, req.OptionID, req.Author, req.Body)

	c, err := scanComment(row)
	if err != nil {
		return Comment{}, fmt.Errorf("create comment: %w", err)
	}
	return *c, nil
}

func (r *pgRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM mission_comments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanComment(s rowScanner) (*Comment, error) {
	var c Comment
	err := s.Scan(
		&c.ID, &c.MissionID, &c.OptionID,
		&c.Author, &c.Body,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan comment: %w", err)
	}
	return &c, nil
}

// IsNotFound reports whether err is a pgx not-found error.
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
