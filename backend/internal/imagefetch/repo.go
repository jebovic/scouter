package imagefetch

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type InsertParams struct {
	OptionID    uuid.UUID
	MinioKey    string
	ContentType string
	Width       *int
	Height      *int
	SourceURL   string
	SortOrder   int
}

func (r *Repository) Insert(ctx context.Context, p InsertParams) (*OptionImage, error) {
	row := r.pool.QueryRow(ctx, `
		INSERT INTO option_images (option_id, minio_key, content_type, width, height, source_url, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at`,
		p.OptionID, p.MinioKey, p.ContentType, p.Width, p.Height, p.SourceURL, p.SortOrder,
	)
	return scanImage(row)
}

func (r *Repository) ListByOption(ctx context.Context, optionID uuid.UUID) ([]*OptionImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at
		FROM option_images WHERE option_id = $1 ORDER BY sort_order, created_at`,
		optionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	defer rows.Close()
	var out []*OptionImage
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

func (r *Repository) SourceURLExists(ctx context.Context, optionID uuid.UUID, sourceURL string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM option_images WHERE option_id=$1 AND source_url=$2)`,
		optionID, sourceURL,
	).Scan(&exists)
	return exists, err
}

func (r *Repository) UpdateSortOrder(ctx context.Context, imageID uuid.UUID, sortOrder int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE option_images SET sort_order=$1 WHERE id=$2`,
		sortOrder, imageID,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, imageID uuid.UUID) (*OptionImage, error) {
	row := r.pool.QueryRow(ctx, `
		DELETE FROM option_images WHERE id=$1
		RETURNING id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at`,
		imageID,
	)
	return scanImage(row)
}

func (r *Repository) ListAll(ctx context.Context) ([]*OptionImage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at
		FROM option_images ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list all images: %w", err)
	}
	defer rows.Close()
	var out []*OptionImage
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, img)
	}
	return out, rows.Err()
}

// DeleteByKey removes the option_images row with the given minio_key.
func (r *Repository) DeleteByKey(ctx context.Context, minioKey string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM option_images WHERE minio_key=$1`, minioKey)
	return err
}

type scanner interface{ Scan(dest ...any) error }

func scanImage(s scanner) (*OptionImage, error) {
	var img OptionImage
	err := s.Scan(
		&img.ID, &img.OptionID, &img.MinioKey, &img.ContentType,
		&img.Width, &img.Height, &img.SourceURL, &img.SortOrder, &img.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan image: %w", err)
	}
	return &img, nil
}
