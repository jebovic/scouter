package option

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for options.
type Repository interface {
	ListByMission(ctx context.Context, missionID uuid.UUID) ([]Option, error)
	ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]Option, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Option, error)
	Create(ctx context.Context, o Option) (*Option, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Option, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMission(ctx context.Context, missionID uuid.UUID) error
	Pin(ctx context.Context, id uuid.UUID) (*Option, error)
	Reject(ctx context.Context, id uuid.UUID, req RejectRequest) (*Option, error)
	Unreject(ctx context.Context, id uuid.UUID) (*Option, error)
	DeletePinned(ctx context.Context, missionID uuid.UUID) error
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed option repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const selectCols = `id, mission_id, name, category, badge, attributes, price_range, notes, warnings, url, pinned, rejected, reject_reason, created_at`

func (r *pgRepository) ListByMission(ctx context.Context, missionID uuid.UUID) ([]Option, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM options WHERE mission_id = $1 ORDER BY created_at ASC`, missionID)
	if err != nil {
		return nil, fmt.Errorf("list options: %w", err)
	}
	defer rows.Close()

	var opts []Option
	for rows.Next() {
		o, err := scanOption(rows)
		if err != nil {
			return nil, err
		}
		opts = append(opts, *o)
	}
	return opts, rows.Err()
}

func (r *pgRepository) ListByMissionPaged(ctx context.Context, missionID uuid.UUID, cursor *time.Time, limit int) ([]Option, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM options
		WHERE mission_id = $1 AND ($2::timestamptz IS NULL OR created_at > $2)
		ORDER BY created_at ASC
		LIMIT $3`, missionID, cursor, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list options paged: %w", err)
	}
	defer rows.Close()

	var opts []Option
	for rows.Next() {
		o, err := scanOption(rows)
		if err != nil {
			return nil, err
		}
		opts = append(opts, *o)
	}
	return opts, rows.Err()
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*Option, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectCols+`
		FROM options WHERE id = $1`, id)
	o, err := scanOption(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}

func (r *pgRepository) Create(ctx context.Context, o Option) (*Option, error) {
	attrsJSON, err := json.Marshal(o.Attributes)
	if err != nil {
		return nil, fmt.Errorf("marshal attributes: %w", err)
	}
	warningsJSON, err := json.Marshal(o.Warnings)
	if err != nil {
		return nil, fmt.Errorf("marshal warnings: %w", err)
	}

	var priceRangeJSON []byte
	if o.PriceRange != nil {
		priceRangeJSON, err = json.Marshal(o.PriceRange)
		if err != nil {
			return nil, fmt.Errorf("marshal price_range: %w", err)
		}
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO options (mission_id, name, category, badge, attributes, price_range, notes, warnings, url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+selectCols,
		o.MissionID, o.Name, o.Category, o.Badge,
		attrsJSON, priceRangeJSON, o.Notes, warningsJSON, o.URL)

	return scanOption(row)
}

func (r *pgRepository) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Option, error) {
	var attrsJSON, warningsJSON, priceRangeJSON []byte
	var err error

	if req.Attributes != nil {
		attrsJSON, err = json.Marshal(req.Attributes)
		if err != nil {
			return nil, fmt.Errorf("marshal attributes: %w", err)
		}
	}
	if req.Warnings != nil {
		warningsJSON, err = json.Marshal(req.Warnings)
		if err != nil {
			return nil, fmt.Errorf("marshal warnings: %w", err)
		}
	}
	if req.PriceRange != nil {
		priceRangeJSON, err = json.Marshal(req.PriceRange)
		if err != nil {
			return nil, fmt.Errorf("marshal price_range: %w", err)
		}
	}

	row := r.pool.QueryRow(ctx, `
		UPDATE options SET
		  badge       = COALESCE($2, badge),
		  attributes  = CASE WHEN $3::jsonb IS NOT NULL THEN $3 ELSE attributes END,
		  price_range = CASE WHEN $4::jsonb IS NOT NULL THEN $4 ELSE price_range END,
		  notes       = COALESCE($5, notes),
		  warnings    = CASE WHEN $6::jsonb IS NOT NULL THEN $6 ELSE warnings END
		WHERE id = $1
		RETURNING `+selectCols,
		id, req.Badge, attrsJSON, priceRangeJSON, req.Notes, warningsJSON)

	o, err := scanOption(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM options WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete option: %w", err)
	}
	return nil
}

// DeleteByMission deletes non-pinned, non-rejected options for a mission.
// Pinned and rejected rows are preserved as positive/negative examples for the LLM.
func (r *pgRepository) DeleteByMission(ctx context.Context, missionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM options WHERE mission_id = $1 AND pinned = false AND rejected = false`, missionID)
	if err != nil {
		return fmt.Errorf("delete options by mission: %w", err)
	}
	return nil
}

// Pin toggles the pinned flag. When pinning (false→true) it also clears the rejected flag.
func (r *pgRepository) Pin(ctx context.Context, id uuid.UUID) (*Option, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE options SET
		  pinned        = NOT pinned,
		  rejected      = CASE WHEN NOT pinned THEN false      ELSE rejected      END,
		  reject_reason = CASE WHEN NOT pinned THEN NULL       ELSE reject_reason END
		WHERE id = $1
		RETURNING `+selectCols, id)
	o, err := scanOption(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}

func (r *pgRepository) Reject(ctx context.Context, id uuid.UUID, req RejectRequest) (*Option, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE options SET rejected = true, reject_reason = $2, pinned = false WHERE id = $1
		RETURNING `+selectCols, id, req.Reason)
	o, err := scanOption(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}

// Unreject clears the rejected flag on an option.
func (r *pgRepository) Unreject(ctx context.Context, id uuid.UUID) (*Option, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE options SET rejected = false, reject_reason = NULL WHERE id = $1
		RETURNING `+selectCols, id)
	o, err := scanOption(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return o, err
}

// DeletePinned removes only pinned options for a mission (e.g. user clears their pin list).
func (r *pgRepository) DeletePinned(ctx context.Context, missionID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM options WHERE mission_id = $1 AND pinned = true`, missionID)
	if err != nil {
		return fmt.Errorf("delete pinned options: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOption(s scanner) (*Option, error) {
	var o Option
	var attrsRaw, warningsRaw, priceRangeRaw []byte
	var rejectReason *string

	err := s.Scan(
		&o.ID, &o.MissionID, &o.Name, &o.Category, &o.Badge,
		&attrsRaw, &priceRangeRaw, &o.Notes, &warningsRaw, &o.URL,
		&o.Pinned, &o.Rejected, &rejectReason, &o.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan option: %w", err)
	}

	if rejectReason != nil {
		o.RejectReason = *rejectReason
	}
	if len(attrsRaw) > 0 {
		if err := json.Unmarshal(attrsRaw, &o.Attributes); err != nil {
			return nil, fmt.Errorf("unmarshal attributes: %w", err)
		}
	}
	if len(warningsRaw) > 0 {
		if err := json.Unmarshal(warningsRaw, &o.Warnings); err != nil {
			return nil, fmt.Errorf("unmarshal warnings: %w", err)
		}
	}
	if len(priceRangeRaw) > 0 {
		o.PriceRange = &PriceRange{}
		if err := json.Unmarshal(priceRangeRaw, o.PriceRange); err != nil {
			return nil, fmt.Errorf("unmarshal price_range: %w", err)
		}
	}

	return &o, nil
}
