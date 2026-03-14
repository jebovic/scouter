// Command seed inserts a sample "Summer Holiday 2026" mission for smoke-testing.
// It is idempotent: if the mission slug already exists it is skipped.
// Usage: DATABASE_URL=... ./scouter-seed
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/jibei/scouter/internal/db"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}

	if err := db.Migrate(dbURL); err != nil {
		log.Error("migration failed", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Error("database connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := seed(ctx, pool, log); err != nil {
		log.Error("seed failed", "err", err)
		os.Exit(1)
	}
	log.Info("seed complete")
}

func seed(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	const slug = "summer-holiday-2026"

	// Idempotency check
	var existing string
	err := pool.QueryRow(ctx, `SELECT slug FROM missions WHERE slug = $1`, slug).Scan(&existing)
	if err == nil {
		log.Info("mission already exists, skipping", "slug", slug)
		return nil
	}
	if err != pgx.ErrNoRows {
		return fmt.Errorf("check existing mission: %w", err)
	}

	constraints := []map[string]any{
		{"key": "max_flight_duration", "label": "Max flight duration", "value": "8h", "type": "hard"},
		{"key": "travel_style", "label": "Travel style", "value": "mix of beach and culture", "type": "soft"},
		{"key": "accommodation", "label": "Accommodation", "value": "4-star hotel or quality Airbnb", "type": "soft"},
	}
	costCategories := []string{"flights", "accommodation", "transport", "activities", "food"}
	timeline := []map[string]any{
		{"date": "2026-06-01", "label": "Booking deadline (best fares)", "urgent": true},
		{"date": "2026-07-15", "label": "Departure target", "urgent": false},
		{"date": "2026-07-29", "label": "Return date", "urgent": false},
	}

	constraintsJSON, err := json.Marshal(constraints)
	if err != nil {
		return fmt.Errorf("marshal constraints: %w", err)
	}
	categoriesJSON, err := json.Marshal(costCategories)
	if err != nil {
		return fmt.Errorf("marshal cost categories: %w", err)
	}
	timelineJSON, err := json.Marshal(timeline)
	if err != nil {
		return fmt.Errorf("marshal timeline: %w", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO missions (slug, name, icon, category, budget, currency, locale, phase, constraints, cost_categories, timeline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		slug, "Summer Holiday 2026", "🏖️", "travel",
		4000.0, "EUR", "fr-FR", "researching",
		constraintsJSON, categoriesJSON, timelineJSON,
	)
	if err != nil {
		return fmt.Errorf("insert mission: %w", err)
	}

	log.Info("seeded mission", "slug", slug, "name", "Summer Holiday 2026")
	return nil
}
