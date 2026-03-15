package forecast

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for forecast results.
type Repository interface {
	Save(ctx context.Context, f ForecastResult) (ForecastResult, error)
	GetLatest(ctx context.Context, missionID string) (ForecastResult, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed forecast repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Save(ctx context.Context, f ForecastResult) (ForecastResult, error) {
	recsJSON, err := json.Marshal(f.Recommendations)
	if err != nil {
		return ForecastResult{}, fmt.Errorf("marshal recommendations: %w", err)
	}
	risksJSON, err := json.Marshal(f.RiskFactors)
	if err != nil {
		return ForecastResult{}, fmt.Errorf("marshal risk_factors: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO forecasts
			(mission_id, estimated_total, confidence, recommendations, risk_factors, months_to_save, optimal_buy_time)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)
		RETURNING id, mission_id, estimated_total, confidence, recommendations, risk_factors, months_to_save, optimal_buy_time, created_at`,
		f.MissionID,
		f.EstimatedTotal,
		f.Confidence,
		recsJSON,
		risksJSON,
		f.MonthsToSave,
		f.OptimalBuyTime,
	)

	return scanForecast(row)
}

func (r *pgRepository) GetLatest(ctx context.Context, missionID string) (ForecastResult, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, mission_id, estimated_total, confidence, recommendations, risk_factors, months_to_save, optimal_buy_time, created_at
		FROM forecasts
		WHERE mission_id = $1
		ORDER BY created_at DESC
		LIMIT 1`, missionID)

	f, err := scanForecast(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return ForecastResult{}, ErrNotFound
	}
	return f, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanForecast(s rowScanner) (ForecastResult, error) {
	var f ForecastResult
	var recsRaw, risksRaw []byte

	err := s.Scan(
		&f.ID,
		&f.MissionID,
		&f.EstimatedTotal,
		&f.Confidence,
		&recsRaw,
		&risksRaw,
		&f.MonthsToSave,
		&f.OptimalBuyTime,
		&f.CreatedAt,
	)
	if err != nil {
		return ForecastResult{}, fmt.Errorf("scan forecast: %w", err)
	}

	if len(recsRaw) > 0 {
		if err := json.Unmarshal(recsRaw, &f.Recommendations); err != nil {
			return ForecastResult{}, fmt.Errorf("unmarshal recommendations: %w", err)
		}
	}
	if f.Recommendations == nil {
		f.Recommendations = []string{}
	}

	if len(risksRaw) > 0 {
		if err := json.Unmarshal(risksRaw, &f.RiskFactors); err != nil {
			return ForecastResult{}, fmt.Errorf("unmarshal risk_factors: %w", err)
		}
	}
	if f.RiskFactors == nil {
		f.RiskFactors = []string{}
	}

	return f, nil
}
