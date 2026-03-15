package wishlist

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const wishlistCols = `id, name, url, target_price, currency, notes, created_at, updated_at,
	last_checked_at, last_checked_price, alert_enabled`

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func scanItem(s interface {
	Scan(...any) error
}) (WishListItem, error) {
	var item WishListItem
	err := s.Scan(
		&item.ID, &item.Name, &item.URL, &item.TargetPrice, &item.Currency, &item.Notes,
		&item.CreatedAt, &item.UpdatedAt,
		&item.LastCheckedAt, &item.LastCheckedPrice, &item.AlertEnabled,
	)
	return item, err
}

func (r *Repository) Create(ctx context.Context, req CreateWishListItemRequest) (WishListItem, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO wish_list_items (name, url, target_price, currency, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING `+wishlistCols,
		req.Name, req.URL, req.TargetPrice, req.Currency, req.Notes,
	)
	return scanItem(row)
}

func (r *Repository) List(ctx context.Context) ([]WishListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+wishlistCols+`
		FROM wish_list_items
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []WishListItem{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM wish_list_items WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (WishListItem, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+wishlistCols+`
		FROM wish_list_items WHERE id = $1
	`, id)
	item, err := scanItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return WishListItem{}, pgx.ErrNoRows
	}
	return item, err
}

// ListAlertCandidates returns items with alert_enabled=true and a target_price set,
// that have either never been checked or were last checked more than 24 hours ago.
func (r *Repository) ListAlertCandidates(ctx context.Context) ([]WishListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+wishlistCols+`
		FROM wish_list_items
		WHERE alert_enabled = true
		  AND target_price IS NOT NULL
		  AND (last_checked_at IS NULL OR last_checked_at < NOW() - INTERVAL '24 hours')
		ORDER BY last_checked_at ASC NULLS FIRST
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []WishListItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// UpdateAlertEnabled toggles the alert_enabled flag for a wish list item.
func (r *Repository) UpdateAlertEnabled(ctx context.Context, id string, enabled bool) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE wish_list_items SET alert_enabled = $1, updated_at = NOW() WHERE id = $2`,
		enabled, id,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

// UpdatePriceCheck records the latest LLM price estimate for an item.
func (r *Repository) UpdatePriceCheck(ctx context.Context, id string, price float64) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE wish_list_items
		SET last_checked_at = NOW(), last_checked_price = $1, updated_at = NOW()
		WHERE id = $2
	`, price, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
