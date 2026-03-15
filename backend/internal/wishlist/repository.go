package wishlist

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, req CreateWishListItemRequest) (WishListItem, error) {
	var item WishListItem
	err := r.pool.QueryRow(ctx, `
		INSERT INTO wish_list_items (name, url, target_price, currency, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, url, target_price, currency, notes, created_at, updated_at
	`, req.Name, req.URL, req.TargetPrice, req.Currency, req.Notes).Scan(
		&item.ID, &item.Name, &item.URL, &item.TargetPrice, &item.Currency, &item.Notes,
		&item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}

func (r *Repository) List(ctx context.Context) ([]WishListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, url, target_price, currency, notes, created_at, updated_at
		FROM wish_list_items
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []WishListItem{}
	for rows.Next() {
		var item WishListItem
		if err := rows.Scan(&item.ID, &item.Name, &item.URL, &item.TargetPrice, &item.Currency, &item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
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
	var item WishListItem
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, url, target_price, currency, notes, created_at, updated_at
		FROM wish_list_items WHERE id = $1
	`, id).Scan(&item.ID, &item.Name, &item.URL, &item.TargetPrice, &item.Currency, &item.Notes, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return WishListItem{}, pgx.ErrNoRows
	}
	return item, err
}
