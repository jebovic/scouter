package receipt

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is the data access interface for receipt analyses.
type Repository interface {
	Save(ctx context.Context, a ReceiptAnalysis) (*ReceiptAnalysis, error)
	ListByMission(ctx context.Context, slug string) ([]ReceiptAnalysis, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed receipt repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

// Save persists a new ReceiptAnalysis and returns the saved record with DB-generated fields.
func (r *pgRepository) Save(ctx context.Context, a ReceiptAnalysis) (*ReceiptAnalysis, error) {
	itemsJSON, err := json.Marshal(a.Items)
	if err != nil {
		return nil, fmt.Errorf("marshal receipt items: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO receipt_analyses (mission_slug, raw_text, items, total, currency)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, mission_slug, raw_text, items, total, currency, analyzed_at`,
		a.MissionSlug, a.RawText, itemsJSON, a.Total, a.Currency,
	)

	return scanAnalysis(row)
}

// ListByMission returns the last 10 analyses for the given mission slug, newest first.
func (r *pgRepository) ListByMission(ctx context.Context, slug string) ([]ReceiptAnalysis, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, mission_slug, raw_text, items, total, currency, analyzed_at
		FROM receipt_analyses
		WHERE mission_slug = $1
		ORDER BY analyzed_at DESC
		LIMIT 10`,
		slug,
	)
	if err != nil {
		return nil, fmt.Errorf("list receipt analyses: %w", err)
	}
	defer rows.Close()

	var analyses []ReceiptAnalysis
	for rows.Next() {
		a, scanErr := scanAnalysisRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		analyses = append(analyses, *a)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list receipt analyses rows: %w", rows.Err())
	}
	if analyses == nil {
		analyses = []ReceiptAnalysis{}
	}
	return analyses, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAnalysis(s rowScanner) (*ReceiptAnalysis, error) {
	var a ReceiptAnalysis
	var itemsJSON []byte
	err := s.Scan(
		&a.ID, &a.MissionSlug, &a.RawText,
		&itemsJSON, &a.Total, &a.Currency, &a.AnalyzedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan receipt analysis: %w", err)
	}
	if err := json.Unmarshal(itemsJSON, &a.Items); err != nil {
		return nil, fmt.Errorf("unmarshal receipt items: %w", err)
	}
	if a.Items == nil {
		a.Items = []ReceiptItem{}
	}
	return &a, nil
}

// scanAnalysisRow wraps pgx rows into the rowScanner interface for scanAnalysis.
func scanAnalysisRow(rows interface{ Scan(...any) error }) (*ReceiptAnalysis, error) {
	return scanAnalysis(rows)
}
