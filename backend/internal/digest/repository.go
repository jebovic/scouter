package digest

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository queries price_history to find items whose price dropped
// compared to their price N days ago.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a Repository backed by pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// GetPriceDrops returns shopping items that cost less now than they did
// `days` days ago, across all active (or recently archived) missions.
// Results are ordered by drop percentage descending.
func (r *Repository) GetPriceDrops(ctx context.Context, days int) ([]PriceDropItem, error) {
	query := fmt.Sprintf(`
WITH latest AS (
  SELECT DISTINCT ON (item_id) item_id, price, recorded_at
  FROM price_history
  ORDER BY item_id, recorded_at DESC
),
week_ago AS (
  SELECT DISTINCT ON (item_id) item_id, price
  FROM price_history
  WHERE recorded_at < NOW() - INTERVAL '%d days'
  ORDER BY item_id, recorded_at DESC
)
SELECT
  si.id::text        AS item_id,
  si.name            AS item_name,
  COALESCE(si.merchant, '') AS merchant,
  COALESCE(si.currency, 'EUR') AS currency,
  w.price            AS price_before,
  l.price            AS price_now,
  m.id::text         AS mission_id,
  m.name             AS mission_name,
  m.slug             AS mission_slug
FROM shopping_items si
JOIN latest  l ON l.item_id = si.id
JOIN week_ago w ON w.item_id = si.id
JOIN missions m ON m.id = si.mission_id
WHERE l.price < w.price
  AND (m.archived_at IS NULL OR m.archived_at > NOW() - INTERVAL '30 days')
ORDER BY (w.price - l.price) / w.price DESC
`, days)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PriceDropItem
	for rows.Next() {
		var it PriceDropItem
		if err := rows.Scan(
			&it.ItemID,
			&it.ItemName,
			&it.Merchant,
			&it.Currency,
			&it.PriceBefore,
			&it.PriceNow,
			&it.MissionID,
			&it.MissionName,
			&it.MissionSlug,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
