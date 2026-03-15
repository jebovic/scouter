// Command seed inserts sample missions for smoke-testing and E2E development.
// It is idempotent: missions that already exist (by slug) are skipped.
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

	if err := seedHomeServer(ctx, pool, log); err != nil {
		log.Error("seed home server failed", "err", err)
		os.Exit(1)
	}
	if err := seedSummerHoliday(ctx, pool, log); err != nil {
		log.Error("seed summer holiday failed", "err", err)
		os.Exit(1)
	}
	log.Info("seed complete")
}

// seedHomeServer creates the "home-server-2026" mission with options and shopping items.
// This is the primary fixture mission for E2E tests.
func seedHomeServer(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	const slug = "home-server-2026"

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
		{"key": "noise_level", "label": "Noise level", "value": "Silent or near-silent (living space)", "type": "hard"},
		{"key": "idle_power", "label": "Idle power draw", "value": "Under 50W at idle", "type": "hard"},
		{"key": "form_factor", "label": "Form factor", "value": "Mini-ITX or NAS enclosure preferred", "type": "soft"},
		{"key": "storage_bays", "label": "Storage bays", "value": "At least 4 drive bays for future expansion", "type": "soft"},
	}
	costCategories := []string{"cpu_motherboard", "memory", "storage", "case_psu", "networking"}
	timeline := []map[string]any{
		{"date": "2026-04-15", "label": "Components research deadline", "urgent": false},
		{"date": "2026-05-01", "label": "Purchase target (spring sales)", "urgent": true},
		{"date": "2026-06-01", "label": "Setup and configuration done", "urgent": false},
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

	var missionID string
	err = pool.QueryRow(ctx, `
		INSERT INTO missions (slug, name, icon, category, budget, currency, locale, phase, constraints, cost_categories, timeline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`,
		slug, "Home Server Build", "🖥️", "computing",
		2000.0, "EUR", "fr-FR", "comparing",
		constraintsJSON, categoriesJSON, timelineJSON,
	).Scan(&missionID)
	if err != nil {
		return fmt.Errorf("insert mission: %w", err)
	}
	log.Info("seeded mission", "slug", slug, "id", missionID)

	if err := seedHomeServerOptions(ctx, pool, log, missionID); err != nil {
		return fmt.Errorf("seed options: %w", err)
	}
	if err := seedHomeServerShoppingItems(ctx, pool, log, missionID); err != nil {
		return fmt.Errorf("seed shopping items: %w", err)
	}
	return nil
}

func seedHomeServerOptions(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger, missionID string) error {
	type optionRow struct {
		name       string
		category   string
		badge      string
		attributes []map[string]any
		priceRange map[string]any
		notes      string
		warnings   []string
		url        string
		rejected   bool
		rejectRsn  string
	}

	options := []optionRow{
		{
			name:     "Synology DS923+",
			category: "NAS",
			badge:    "recommended",
			attributes: []map[string]any{
				{"key": "cpu", "label": "CPU", "value": "AMD Ryzen R1600 dual-core 2.6 GHz", "type": "text"},
				{"key": "ram", "label": "RAM", "value": "4 GB DDR4 ECC (expandable to 32 GB)", "type": "text"},
				{"key": "bays", "label": "Drive bays", "value": "4 (expandable to 9 with DX517)", "type": "text"},
				{"key": "idle_power", "label": "Idle power", "value": "~19W (HDD off)", "type": "text"},
				{"key": "os", "label": "OS", "value": "Synology DSM 7.2", "type": "text"},
				{"key": "noise", "label": "Noise", "value": "21.7 dB(A) — near silent", "type": "text"},
				{"key": "price_score", "label": "Value score", "value": 78, "type": "score", "max": 100},
			},
			priceRange: map[string]any{"min": 580.0, "max": 650.0, "best": 609.0},
			notes:      "Best all-in-one NAS for home use. Excellent DSM ecosystem, apps, and mobile integrations. Drive-only cost is extra (~€320 for 2×4TB WD Red Plus). Total budget: ~€930.",
			warnings:   []string{"Drives sold separately", "Proprietary OS limits custom software"},
			url:        "https://www.synology.com/fr-fr/products/DS923+",
		},
		{
			name:     "Custom mini-ITX — Intel N100",
			category: "Custom build",
			badge:    "alternative",
			attributes: []map[string]any{
				{"key": "cpu", "label": "CPU", "value": "Intel N100 (4C/4T, 3.4 GHz burst, 6W TDP)", "type": "text"},
				{"key": "ram", "label": "RAM", "value": "32 GB DDR4 SO-DIMM", "type": "text"},
				{"key": "bays", "label": "Drive bays", "value": "6 (Fractal Node 304)", "type": "text"},
				{"key": "idle_power", "label": "Idle power", "value": "~8–12W with SSD, ~18W with HDDs", "type": "text"},
				{"key": "os", "label": "OS", "value": "TrueNAS SCALE / Proxmox + TrueNAS VM", "type": "text"},
				{"key": "noise", "label": "Noise", "value": "Near-silent with Noctua fan swap", "type": "text"},
				{"key": "price_score", "label": "Value score", "value": 88, "type": "score", "max": 100},
			},
			priceRange: map[string]any{"min": 380.0, "max": 520.0, "best": 430.0},
			notes:      "Maximum flexibility at lowest cost. Run VMs, containers, and Plex on the same box. Full Linux ecosystem. Requires more setup time vs Synology.",
			warnings:   []string{"Requires manual OS setup and maintenance", "No vendor support"},
			url:        "https://ark.intel.com/content/www/fr/fr/ark/products/231803/intel-processor-n100-6m-cache-up-to-3-40-ghz.html",
		},
		{
			name:     "Raspberry Pi 5 — 8 GB",
			category: "SBC",
			badge:    "watch",
			attributes: []map[string]any{
				{"key": "cpu", "label": "CPU", "value": "Broadcom BCM2712 quad-core A76 @ 2.4 GHz", "type": "text"},
				{"key": "ram", "label": "RAM", "value": "8 GB LPDDR4X", "type": "text"},
				{"key": "bays", "label": "Drive bays", "value": "0 (external USB / NVMe HAT required)", "type": "text"},
				{"key": "idle_power", "label": "Idle power", "value": "~3–5W", "type": "text"},
				{"key": "os", "label": "OS", "value": "Raspberry Pi OS / DietPi", "type": "text"},
				{"key": "noise", "label": "Noise", "value": "Silent (no fan needed with heatsink case)", "type": "text"},
				{"key": "price_score", "label": "Value score", "value": 60, "type": "score", "max": 100},
			},
			priceRange: map[string]any{"min": 80.0, "max": 160.0, "best": 95.0},
			notes:      "Ultra-low power and cost. Best for lightweight file serving and Pihole. Not suited for heavy transcoding or large RAID arrays without add-on HATs.",
			warnings:   []string{"Limited PCIe bandwidth for multi-drive setups", "No ECC memory", "USB 3.0 bus shared with networking on older hats"},
			url:        "https://www.raspberrypi.com/products/raspberry-pi-5/",
		},
		{
			name:      "Dell PowerEdge T150",
			category:  "Tower server",
			badge:     "rejected",
			rejected:  true,
			rejectRsn: "Too loud (45 dB+) and power-hungry (80–200W) for a home living space. Overkill for personal NAS use case.",
			attributes: []map[string]any{
				{"key": "cpu", "label": "CPU", "value": "Intel Xeon E-2300 series", "type": "text"},
				{"key": "ram", "label": "RAM", "value": "Up to 128 GB ECC DDR4", "type": "text"},
				{"key": "bays", "label": "Drive bays", "value": "4 × 3.5\" LFF", "type": "text"},
				{"key": "idle_power", "label": "Idle power", "value": "~80–120W", "type": "text"},
				{"key": "noise", "label": "Noise", "value": "45+ dB(A) — too loud", "type": "text"},
				{"key": "price_score", "label": "Value score", "value": 35, "type": "score", "max": 100},
			},
			priceRange: map[string]any{"min": 650.0, "max": 950.0, "best": 720.0},
			notes:      "Enterprise-grade reliability but wrong tool for a quiet home server. High power cost (~€150/year extra vs N100 build).",
			warnings:   []string{"Very loud fans", "High electricity cost", "No iGPU for media transcoding"},
			url:        "https://www.dell.com/fr-fr/shop/servers-storage-and-networking/poweredge-t150-tower-server/spd/poweredge-t150",
		},
	}

	for _, o := range options {
		attrJSON, err := json.Marshal(o.attributes)
		if err != nil {
			return fmt.Errorf("marshal attributes for %s: %w", o.name, err)
		}
		priceJSON, err := json.Marshal(o.priceRange)
		if err != nil {
			return fmt.Errorf("marshal price range for %s: %w", o.name, err)
		}
		warnJSON, err := json.Marshal(o.warnings)
		if err != nil {
			return fmt.Errorf("marshal warnings for %s: %w", o.name, err)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO options (mission_id, name, category, badge, attributes, price_range, notes, warnings, url, rejected, reject_reason)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			missionID, o.name, o.category, o.badge,
			attrJSON, priceJSON, o.notes, warnJSON, o.url, o.rejected, o.rejectRsn,
		)
		if err != nil {
			return fmt.Errorf("insert option %s: %w", o.name, err)
		}
		log.Info("seeded option", "name", o.name)
	}
	return nil
}

func seedHomeServerShoppingItems(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger, missionID string) error {
	type itemRow struct {
		name            string
		merchant        string
		costCategory    string
		price           float64
		originalEst     float64
		targetPrice     float64
		status          string
		note            string
		url             string
	}

	items := []itemRow{
		{
			name:         "ASRock N100DC-ITX (board + N100 combo)",
			merchant:     "LDLC",
			costCategory: "cpu_motherboard",
			price:        149.90,
			originalEst:  140.0,
			targetPrice:  130.0,
			status:       "buy",
			note:         "Fanless DC-in board. Pairs with a PicoPSU for ultra-quiet build. In stock.",
			url:          "https://www.ldlc.com/fiche/PB00565862.html",
		},
		{
			name:         "Kingston Fury Beast 32 GB DDR4-3200 SO-DIMM",
			merchant:     "Amazon FR",
			costCategory: "memory",
			price:        72.99,
			originalEst:  80.0,
			targetPrice:  65.0,
			status:       "buy",
			note:         "2×16 GB kit. N100 supports dual-channel DDR4 up to 3200 MT/s.",
			url:          "https://www.amazon.fr/dp/B09VDMG9PC",
		},
		{
			name:         "Samsung 870 EVO 2 TB SATA SSD (×2)",
			merchant:     "Materiel.net",
			costCategory: "storage",
			price:        209.98,
			originalEst:  240.0,
			targetPrice:  185.0,
			status:       "watch",
			note:         "Pair for ZFS mirror (RAID-1 equivalent). Price often drops during Amazon Prime Day.",
			url:          "https://www.materiel.net/produit/202104160001.html",
		},
		{
			name:         "WD Red Plus 4 TB 3.5\" CMR (×2)",
			merchant:     "Rueducommerce",
			costCategory: "storage",
			price:        174.00,
			originalEst:  200.0,
			targetPrice:  150.0,
			status:       "watch",
			note:         "NAS-rated CMR drives for bulk cold storage. Avoid SMR WD Red (non-Plus) for ZFS.",
			url:          "https://www.rueducommerce.fr/produit/wd-red-plus-4-to",
		},
		{
			name:         "Fractal Design Node 304",
			merchant:     "LDLC",
			costCategory: "case_psu",
			price:        74.90,
			originalEst:  75.0,
			targetPrice:  65.0,
			status:       "buy",
			note:         "6 × 3.5\" hot-swap bays, mini-ITX, near-silent fan. Best compact NAS case available.",
			url:          "https://www.ldlc.com/fiche/PB00142440.html",
		},
		{
			name:         "Corsair SF450 SFX PSU 80+ Gold",
			merchant:     "Amazon FR",
			costCategory: "case_psu",
			price:        89.99,
			originalEst:  95.0,
			targetPrice:  80.0,
			status:       "watch",
			note:         "Overkill for N100 (~70W peak) but excellent efficiency curve at low loads. Alternatively use 120W PicoPSU.",
			url:          "https://www.amazon.fr/dp/B07VPHMS5W",
		},
		{
			name:         "TP-Link TL-SG108E 8-port Managed Switch",
			merchant:     "Amazon FR",
			costCategory: "networking",
			price:        29.99,
			originalEst:  35.0,
			targetPrice:  25.0,
			status:       "defer",
			note:         "Nice-to-have for VLAN segmentation. Not needed for initial setup — defer to phase 2.",
			url:          "https://www.amazon.fr/dp/B00K4DS5KU",
		},
	}

	for _, item := range items {
		_, err := pool.Exec(ctx, `
			INSERT INTO shopping_items (mission_id, name, merchant, cost_category, price, original_estimate, target_price, status, note, url)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			missionID, item.name, item.merchant, item.costCategory,
			item.price, item.originalEst, item.targetPrice,
			item.status, item.note, item.url,
		)
		if err != nil {
			return fmt.Errorf("insert shopping item %s: %w", item.name, err)
		}
		log.Info("seeded shopping item", "name", item.name)
	}
	return nil
}

// seedSummerHoliday creates the original "summer-holiday-2026" sample mission.
func seedSummerHoliday(ctx context.Context, pool *pgxpool.Pool, log *slog.Logger) error {
	const slug = "summer-holiday-2026"

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
