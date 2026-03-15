package settings

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the storage interface for settings.
type Repository interface {
	GetAll(ctx context.Context) (map[string]json.RawMessage, error)
	Set(ctx context.Context, key string, value json.RawMessage) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository returns a PostgreSQL-backed Repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// GetAll returns all settings as a map of key -> raw JSON value.
func (r *pgRepository) GetAll(ctx context.Context) (map[string]json.RawMessage, error) {
	rows, err := r.pool.Query(ctx, `SELECT key, value FROM settings ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]json.RawMessage)
	for rows.Next() {
		var key string
		var value json.RawMessage
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, rows.Err()
}

// Set upserts a single setting by key.
func (r *pgRepository) Set(ctx context.Context, key string, value json.RawMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, key, value)
	return err
}
