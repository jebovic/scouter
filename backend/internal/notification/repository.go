package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for notifications.
type Repository interface {
	List(ctx context.Context, limit int) ([]Notification, error)
	UnreadCount(ctx context.Context) (int, error)
	Create(ctx context.Context, req CreateRequest) (*Notification, error)
	MarkRead(ctx context.Context, id uuid.UUID) error
	MarkAllRead(ctx context.Context) error
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

func (r *pgRepository) UnreadCount(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE read = false`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("unread count: %w", err)
	}
	return count, nil
}

func (r *pgRepository) Create(ctx context.Context, req CreateRequest) (*Notification, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (mission_id, item_id, type, title, body)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+notifCols,
		req.MissionID, req.ItemID, req.Type, req.Title, req.Body)
	return scanNotification(row)
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
