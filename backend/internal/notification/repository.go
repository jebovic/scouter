package notification

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ListFilter holds optional query filters for ListFiltered.
type ListFilter struct {
	MissionID *uuid.UUID
	Type      *Type
	Limit     int
}

// Repository defines the data access interface for notifications.
type Repository interface {
	List(ctx context.Context, limit int) ([]Notification, error)
	ListFiltered(ctx context.Context, f ListFilter) ([]Notification, error)
	UnreadCount(ctx context.Context) (int, error)
	Create(ctx context.Context, req CreateRequest) (*Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkAllRead(ctx context.Context) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed notification repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const notifCols = `id, mission_id, item_id, type, title, body, read, created_at`

func (r *pgRepository) List(ctx context.Context, limit int) ([]Notification, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+notifCols+`
		FROM notifications
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var notifs []Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifs = append(notifs, *n)
	}
	return notifs, rows.Err()
}

func (r *pgRepository) ListFiltered(ctx context.Context, f ListFilter) ([]Notification, error) {
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}

	// Build query dynamically based on filters.
	args := []any{limit}
	where := ""
	if f.MissionID != nil {
		args = append(args, *f.MissionID)
		where += fmt.Sprintf(" AND mission_id = $%d", len(args))
	}
	if f.Type != nil {
		args = append(args, string(*f.Type))
		where += fmt.Sprintf(" AND type = $%d", len(args))
	}

	q := `SELECT ` + notifCols + ` FROM notifications WHERE true` + where + ` ORDER BY created_at DESC LIMIT $1`
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list filtered notifications: %w", err)
	}
	defer rows.Close()

	var notifs []Notification
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		notifs = append(notifs, *n)
	}
	return notifs, rows.Err()
}

func (r *pgRepository) UnreadCount(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE read = false`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

func (r *pgRepository) Create(ctx context.Context, req CreateRequest) (*Notification, error) {
	// ON CONFLICT DO NOTHING implements the dedup guard enforced by
	// notifications_dedup_idx (one item+type per day). Returns nil when
	// the notification was a duplicate (not an error).
	row := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (mission_id, item_id, type, title, body)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (item_id, type, (created_at::date)) WHERE item_id IS NOT NULL
		DO NOTHING
		RETURNING `+notifCols,
		req.MissionID, req.ItemID, req.Type, req.Title, req.Body)
	n, err := scanNotification(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // duplicate — silently ignored
		}
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return n, nil
}

func (r *pgRepository) MarkRead(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `UPDATE notifications SET read = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *pgRepository) MarkAllRead(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET read = true WHERE read = false`)
	if err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

// Delete removes a notification by ID. Returns pgx.ErrNoRows if not found.
func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM notifications WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(s scanner) (*Notification, error) {
	var n Notification
	err := s.Scan(&n.ID, &n.MissionID, &n.ItemID, &n.Type, &n.Title, &n.Body, &n.Read, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan notification: %w", err)
	}
	return &n, nil
}
