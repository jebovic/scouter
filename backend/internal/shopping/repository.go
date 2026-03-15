package shopping

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for shopping items and price history.
type Repository interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]Item, error)
	ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]Item, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
	Create(ctx context.Context, item Item) (*Item, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Item, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMission(ctx context.Context, missionID uuid.UUID) error
	AddPriceSnapshot(ctx context.Context, itemID uuid.UUID, req PriceSnapshotRequest) (*PriceSnapshot, error)
	ListPriceHistory(ctx context.Context, itemID uuid.UUID) ([]PriceSnapshot, error)
	// GetPriceHistory returns the last `limit` price observations for an item, ordered ASC by recorded_at.
	// It maps the `note` column to PriceHistoryPoint.Source for charting consumers.
	GetPriceHistory(ctx context.Context, itemID uuid.UUID, limit int) ([]PriceHistoryPoint, error)
	Pin(ctx context.Context, id uuid.UUID) (*Item, error)
	DeletePinned(ctx context.Context, missionID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed shopping repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const selectCols = `id, mission_id, name, merchant, cost_category, price, original_estimate, target_price, status, note, url, pinned, created_at`

func (r *pgRepository) ListByMission(ctx context.Context, missionID uuid.UUID) ([]Item, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM shopping_items WHERE mission_id = $1 ORDER BY cost_category, created_at ASC`, missionID)
	if err != nil {
		return nil, fmt.Errorf("list shopping items: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *pgRepository) ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]Item, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM shopping_items
		WHERE mission_id = $1 AND ($2::timestamptz IS NULL OR created_at > $2)
		ORDER BY created_at ASC
		LIMIT $3`, missionID, cursor, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list shopping items paged: %w", err)
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectCols+`
		FROM shopping_items WHERE id = $1`, id)
	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

func (r *pgRepository) Create(ctx context.Context, item Item) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO shopping_items (mission_id, name, merchant, cost_category, price, original_estimate, target_price, status, note, url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING `+selectCols,
		item.MissionID, item.Name, item.Merchant, item.CostCategory,
		item.Price, item.OriginalEstimate, item.TargetPrice, item.Status, item.Note, item.URL)

	return scanItem(row)
}

func (r *pgRepository) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE shopping_items SET
		  price        = COALESCE($2, price),
		  status       = COALESCE($3, status),
		  note         = COALESCE($4, note),
		  merchant     = COALESCE($5, merchant),
		  target_price = COALESCE($6, target_price)
		WHERE id = $1
		RETURNING `+selectCols,
		id, req.Price, req.Status, req.Note, req.Merchant, req.TargetPrice)

	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM shopping_items WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete shopping item: %w", err)
	}
	return nil
}

// DeleteByMission deletes non-pinned shopping items for a mission.
// Pinned items are preserved as positive examples for the LLM.
func (r *pgRepository) DeleteByMission(ctx context.Context, missionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM shopping_items WHERE mission_id = $1 AND pinned = false`, missionID)
	if err != nil {
		return fmt.Errorf("delete shopping items by mission: %w", err)
	}
	return nil
}

func (r *pgRepository) AddPriceSnapshot(ctx context.Context, itemID uuid.UUID, req PriceSnapshotRequest) (*PriceSnapshot, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO price_history (item_id, price, note)
		VALUES ($1, $2, $3)
		RETURNING id, item_id, price, recorded_at, note`,
		itemID, req.Price, req.Note)

	return scanSnapshot(row)
}

func (r *pgRepository) GetPriceHistory(ctx context.Context, itemID uuid.UUID, limit int) ([]PriceHistoryPoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT price, recorded_at, COALESCE(note, '') AS source
		FROM (
			SELECT price, recorded_at, note
			FROM price_history
			WHERE item_id = $1
			ORDER BY recorded_at DESC
			LIMIT $2
		) sub
		ORDER BY recorded_at ASC`, itemID, limit)
	if err != nil {
		return nil, fmt.Errorf("get price history: %w", err)
	}
	defer rows.Close()

	var pts []PriceHistoryPoint
	for rows.Next() {
		var p PriceHistoryPoint
		if err := rows.Scan(&p.Price, &p.RecordedAt, &p.Source); err != nil {
			return nil, fmt.Errorf("scan price history point: %w", err)
		}
		pts = append(pts, p)
	}
	return pts, rows.Err()
}

func (r *pgRepository) ListPriceHistory(ctx context.Context, itemID uuid.UUID) ([]PriceSnapshot, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, item_id, price, recorded_at, note
		FROM price_history WHERE item_id = $1 ORDER BY recorded_at ASC`, itemID)
	if err != nil {
		return nil, fmt.Errorf("list price history: %w", err)
	}
	defer rows.Close()

	var snapshots []PriceSnapshot
	for rows.Next() {
		s, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, *s)
	}
	return snapshots, rows.Err()
}

// Pin toggles the pinned flag on a shopping item.
func (r *pgRepository) Pin(ctx context.Context, id uuid.UUID) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE shopping_items SET pinned = NOT pinned WHERE id = $1
		RETURNING `+selectCols, id)
	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

// DeletePinned removes only pinned shopping items for a mission.
func (r *pgRepository) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM shopping_items WHERE mission_id = $1 AND pinned = true`, missionID)
	if err != nil {
		return fmt.Errorf("delete pinned shopping items: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanItem(s scanner) (*Item, error) {
	var item Item
	err := s.Scan(
		&item.ID, &item.MissionID, &item.Name, &item.Merchant, &item.CostCategory,
		&item.Price, &item.OriginalEstimate, &item.TargetPrice, &item.Status, &item.Note, &item.URL,
		&item.Pinned, &item.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan shopping item: %w", err)
	}
	return &item, nil
}

func scanSnapshot(s scanner) (*PriceSnapshot, error) {
	var snap PriceSnapshot
	err := s.Scan(&snap.ID, &snap.ItemID, &snap.Price, &snap.RecordedAt, &snap.Note)
	if err != nil {
		return nil, fmt.Errorf("scan price snapshot: %w", err)
	}
	return &snap, nil
}
