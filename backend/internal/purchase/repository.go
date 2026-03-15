package purchase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for purchase records.
type Repository interface {
	Create(ctx context.Context, r PurchaseRecord) (*PurchaseRecord, error)
	GetByMission(ctx context.Context, missionID uuid.UUID) (*PurchaseRecord, error)
	Update(ctx context.Context, missionID uuid.UUID, req UpdateRequest) (*PurchaseRecord, error)
	ExistsByMission(ctx context.Context, missionID uuid.UUID) (bool, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed purchase repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const selectCols = `id, mission_id, option_id, purchased_at, final_price, merchant, satisfaction, review, created_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRecord(row rowScanner) (*PurchaseRecord, error) {
	var r PurchaseRecord
	err := row.Scan(
		&r.ID, &r.MissionID, &r.OptionID,
		&r.PurchasedAt, &r.FinalPrice, &r.Merchant,
		&r.Satisfaction, &r.Review, &r.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (p *pgRepository) Create(ctx context.Context, r PurchaseRecord) (*PurchaseRecord, error) {
	row := p.pool.QueryRow(ctx, `
		INSERT INTO purchase_records (id, mission_id, option_id, purchased_at, final_price, merchant, satisfaction, review)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING `+selectCols,
		r.ID, r.MissionID, r.OptionID, r.PurchasedAt, r.FinalPrice, r.Merchant, r.Satisfaction, r.Review,
	)
	return scanRecord(row)
}

func (p *pgRepository) GetByMission(ctx context.Context, missionID uuid.UUID) (*PurchaseRecord, error) {
	row := p.pool.QueryRow(ctx,
		`SELECT `+selectCols+` FROM purchase_records WHERE mission_id = $1`, missionID)
	rec, err := scanRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

func (p *pgRepository) Update(ctx context.Context, missionID uuid.UUID, req UpdateRequest) (*PurchaseRecord, error) {
	row := p.pool.QueryRow(ctx, `
		UPDATE purchase_records
		SET satisfaction = COALESCE($1, satisfaction),
		    review       = COALESCE($2, review),
		    merchant     = COALESCE($3, merchant),
		    final_price  = COALESCE($4, final_price)
		WHERE mission_id = $5
		RETURNING `+selectCols,
		req.Satisfaction, req.Review, req.Merchant, req.FinalPrice, missionID,
	)
	rec, err := scanRecord(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return rec, err
}

func (p *pgRepository) ExistsByMission(ctx context.Context, missionID uuid.UUID) (bool, error) {
	var exists bool
	err := p.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM purchase_records WHERE mission_id = $1)`, missionID,
	).Scan(&exists)
	return exists, err
}
