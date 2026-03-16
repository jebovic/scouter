package mission

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the data access interface for missions.
type Repository interface {
	List(ctx context.Context) ([]Mission, error)
	ListPaged(ctx context.Context, cursor *time.Time, limit int, includeArchived bool) ([]Mission, error)
	ListActive(ctx context.Context) ([]Mission, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Mission, error)
	GetBySlug(ctx context.Context, slug string) (*Mission, error)
	GetByShareToken(ctx context.Context, token string) (*Mission, error)
	Create(ctx context.Context, m Mission) (*Mission, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Mission, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetPhase(ctx context.Context, id uuid.UUID, phase string) error
	SetShareToken(ctx context.Context, id uuid.UUID, token string) (*Mission, error)
	ClearShareToken(ctx context.Context, id uuid.UUID) error
	Archive(ctx context.Context, id uuid.UUID) (*Mission, error)
	Unarchive(ctx context.Context, id uuid.UUID) (*Mission, error)
	// CloneMission creates a full deep copy of a mission (options + shopping
	// items) in a single transaction. The clone gets a new UUID, a new slug
	// derived from the original (with "-copie" suffix), and the name gets a
	// " (Copie)" suffix. Sensitive fields (lessons, share_token, archived_at)
	// are not copied. Returns the new mission.
	CloneMission(ctx context.Context, slug string) (Mission, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed mission repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

const selectCols = `id, slug, name, icon, category, budget, currency, locale, phase,
		       constraints, cost_categories, timeline, weight_profile,
		       envelope_id, lessons, share_token, archived_at, created_at, updated_at`

func (r *pgRepository) List(ctx context.Context) ([]Mission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM missions ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list missions: %w", err)
	}
	defer rows.Close()

	var missions []Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		missions = append(missions, *m)
	}
	return missions, rows.Err()
}

func (r *pgRepository) ListActive(ctx context.Context) ([]Mission, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+selectCols+`
		FROM missions
		WHERE phase != 'done' AND archived_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list active missions: %w", err)
	}
	defer rows.Close()

	var missions []Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, fmt.Errorf("list active missions: scan: %w", err)
		}
		missions = append(missions, *m)
	}
	return missions, rows.Err()
}

func (r *pgRepository) ListPaged(ctx context.Context, cursor *time.Time, limit int, includeArchived bool) ([]Mission, error) {
	var query string
	if includeArchived {
		query = `
		SELECT ` + selectCols + `
		FROM missions
		WHERE ($1::timestamptz IS NULL OR created_at < $1)
		ORDER BY created_at DESC
		LIMIT $2`
	} else {
		query = `
		SELECT ` + selectCols + `
		FROM missions
		WHERE ($1::timestamptz IS NULL OR created_at < $1)
		  AND archived_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2`
	}
	rows, err := r.pool.Query(ctx, query, cursor, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list missions paged: %w", err)
	}
	defer rows.Close()

	var missions []Mission
	for rows.Next() {
		m, err := scanMission(rows)
		if err != nil {
			return nil, err
		}
		missions = append(missions, *m)
	}
	return missions, rows.Err()
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*Mission, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectCols+`
		FROM missions WHERE id = $1`, id)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug string) (*Mission, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectCols+`
		FROM missions WHERE slug = $1`, slug)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *pgRepository) GetByShareToken(ctx context.Context, token string) (*Mission, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+selectCols+`
		FROM missions WHERE share_token = $1`, token)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *pgRepository) Create(ctx context.Context, m Mission) (*Mission, error) {
	constraintsJSON, err := json.Marshal(m.Constraints)
	if err != nil {
		return nil, fmt.Errorf("marshal constraints: %w", err)
	}
	categoriesJSON, err := json.Marshal(m.CostCategories)
	if err != nil {
		return nil, fmt.Errorf("marshal cost_categories: %w", err)
	}
	timelineJSON, err := json.Marshal(m.Timeline)
	if err != nil {
		return nil, fmt.Errorf("marshal timeline: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO missions (slug, name, icon, category, budget, currency, locale, phase, constraints, cost_categories, timeline, envelope_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING `+selectCols,
		m.Slug, m.Name, m.Icon, m.Category, m.Budget, m.Currency, m.Locale, m.Phase,
		constraintsJSON, categoriesJSON, timelineJSON, m.EnvelopeID)

	return scanMission(row)
}

func (r *pgRepository) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*Mission, error) {
	constraintsJSON, err := json.Marshal(req.Constraints)
	if err != nil {
		return nil, fmt.Errorf("marshal constraints: %w", err)
	}
	categoriesJSON, err := json.Marshal(req.CostCategories)
	if err != nil {
		return nil, fmt.Errorf("marshal cost_categories: %w", err)
	}
	timelineJSON, err := json.Marshal(req.Timeline)
	if err != nil {
		return nil, fmt.Errorf("marshal timeline: %w", err)
	}
	var weightProfileJSON []byte
	if req.WeightProfile != nil {
		weightProfileJSON, err = json.Marshal(req.WeightProfile)
		if err != nil {
			return nil, fmt.Errorf("marshal weight_profile: %w", err)
		}
	}

	row := r.pool.QueryRow(ctx, `
		UPDATE missions SET
		  name            = COALESCE($2, name),
		  icon            = COALESCE($3, icon),
		  budget          = COALESCE($4, budget),
		  phase           = COALESCE($5, phase),
		  lessons         = COALESCE($6, lessons),
		  constraints     = CASE WHEN $7::jsonb IS NOT NULL THEN $7 ELSE constraints END,
		  cost_categories = CASE WHEN $8::jsonb IS NOT NULL THEN $8 ELSE cost_categories END,
		  timeline        = CASE WHEN $9::jsonb IS NOT NULL THEN $9 ELSE timeline END,
		  weight_profile  = CASE WHEN $10::jsonb IS NOT NULL THEN $10 ELSE weight_profile END,
		  envelope_id     = CASE WHEN $11::boolean THEN $12::uuid ELSE envelope_id END,
		  updated_at      = now()
		WHERE id = $1
		RETURNING `+selectCols,
		id, req.Name, req.Icon, req.Budget, req.Phase, req.Lessons,
		string(constraintsJSON), string(categoriesJSON), string(timelineJSON),
		nullableJSON(weightProfileJSON),
		req.EnvelopeID != nil, req.EnvelopeID)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

// nullableJSON returns nil if b is empty or the null literal, otherwise returns the string.
// This allows CASE WHEN $n::jsonb IS NOT NULL to work correctly for optional fields.
func nullableJSON(b []byte) any {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	return string(b)
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM missions WHERE id = $1`, id)
	return err
}

func (r *pgRepository) SetPhase(ctx context.Context, id uuid.UUID, phase string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE missions SET phase = $2, updated_at = now() WHERE id = $1`, id, phase)
	return err
}

func (r *pgRepository) SetShareToken(ctx context.Context, id uuid.UUID, token string) (*Mission, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE missions SET share_token = $2, updated_at = now()
		WHERE id = $1
		RETURNING `+selectCols, id, token)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *pgRepository) ClearShareToken(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE missions SET share_token = NULL, updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *pgRepository) Archive(ctx context.Context, id uuid.UUID) (*Mission, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE missions SET archived_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING `+selectCols, id)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *pgRepository) Unarchive(ctx context.Context, id uuid.UUID) (*Mission, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE missions SET archived_at = NULL, updated_at = now()
		WHERE id = $1
		RETURNING `+selectCols, id)

	m, err := scanMission(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

// CloneMission deep-copies a mission (options + shopping_items) in one transaction.
// The clone gets a new UUID, slug = original_slug + "-copie" (with a short random
// suffix if that slug is already taken), and name = original_name + " (Copie)".
// Sensitive fields (lessons, share_token, archived_at) are not carried over.
func (r *pgRepository) CloneMission(ctx context.Context, slug string) (Mission, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// 1. Fetch original mission.
	row := tx.QueryRow(ctx, `SELECT `+selectCols+` FROM missions WHERE slug = $1`, slug)
	orig, err := scanMission(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Mission{}, fmt.Errorf("mission %q not found", slug)
		}
		return Mission{}, fmt.Errorf("clone mission: fetch original: %w", err)
	}

	// 2. Determine a unique slug for the clone.
	newSlug := slug + "-copie"
	var taken bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM missions WHERE slug = $1)`, newSlug,
	).Scan(&taken); err != nil {
		return Mission{}, fmt.Errorf("clone mission: check slug: %w", err)
	}
	if taken {
		// Append first 4 chars of a new UUID for uniqueness.
		suffix := uuid.New().String()[:4]
		newSlug = slug + "-copie-" + suffix
	}

	// 3. Marshal JSON fields from the original.
	constraintsJSON, err := json.Marshal(orig.Constraints)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: marshal constraints: %w", err)
	}
	categoriesJSON, err := json.Marshal(orig.CostCategories)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: marshal cost_categories: %w", err)
	}
	timelineJSON, err := json.Marshal(orig.Timeline)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: marshal timeline: %w", err)
	}
	var weightProfileJSON []byte
	wp := orig.WeightProfile
	weightProfileJSON, err = json.Marshal(wp)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: marshal weight_profile: %w", err)
	}

	// 4. Insert cloned mission (new UUID assigned by DB default).
	cloneRow := tx.QueryRow(ctx, `
		INSERT INTO missions
		  (slug, name, icon, category, budget, currency, locale, phase,
		   constraints, cost_categories, timeline, weight_profile, envelope_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,'researching',$8,$9,$10,$11,$12)
		RETURNING `+selectCols,
		newSlug,
		orig.Name+" (Copie)",
		orig.Icon,
		orig.Category,
		orig.Budget,
		orig.Currency,
		orig.Locale,
		string(constraintsJSON),
		string(categoriesJSON),
		string(timelineJSON),
		nullableJSON(weightProfileJSON),
		orig.EnvelopeID,
	)
	cloned, err := scanMission(cloneRow)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: insert clone: %w", err)
	}

	// 5. Deep-copy options.
	_, err = tx.Exec(ctx, `
		INSERT INTO options
		  (mission_id, name, category, badge, attributes, price_range, notes, warnings, url)
		SELECT $1, name, category, badge, attributes, price_range, notes, warnings, url
		FROM options
		WHERE mission_id = $2`,
		cloned.ID, orig.ID,
	)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: copy options: %w", err)
	}

	// 6. Deep-copy shopping items (price_history is intentionally not copied —
	//    history belongs to the original purchase context).
	_, err = tx.Exec(ctx, `
		INSERT INTO shopping_items
		  (mission_id, name, merchant, cost_category, price, original_estimate, target_price, status, note, url)
		SELECT $1, name, merchant, cost_category, price, original_estimate, target_price, status, note, url
		FROM shopping_items
		WHERE mission_id = $2`,
		cloned.ID, orig.ID,
	)
	if err != nil {
		return Mission{}, fmt.Errorf("clone mission: copy shopping items: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Mission{}, fmt.Errorf("clone mission: commit: %w", err)
	}

	return *cloned, nil
}

// GenerateShareToken creates a cryptographically random 32-byte base64url token.
func GenerateShareToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate share token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanMission(s scanner) (*Mission, error) {
	var m Mission
	var constraintsRaw, categoriesRaw, timelineRaw, weightProfileRaw []byte

	err := s.Scan(
		&m.ID, &m.Slug, &m.Name, &m.Icon, &m.Category,
		&m.Budget, &m.Currency, &m.Locale, &m.Phase,
		&constraintsRaw, &categoriesRaw, &timelineRaw, &weightProfileRaw,
		&m.EnvelopeID, &m.Lessons, &m.ShareToken, &m.ArchivedAt,
		&m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan mission: %w", err)
	}

	if len(constraintsRaw) > 0 {
		if err := json.Unmarshal(constraintsRaw, &m.Constraints); err != nil {
			return nil, fmt.Errorf("unmarshal constraints: %w", err)
		}
	}
	if len(categoriesRaw) > 0 {
		if err := json.Unmarshal(categoriesRaw, &m.CostCategories); err != nil {
			return nil, fmt.Errorf("unmarshal cost_categories: %w", err)
		}
	}
	if len(timelineRaw) > 0 {
		if err := json.Unmarshal(timelineRaw, &m.Timeline); err != nil {
			return nil, fmt.Errorf("unmarshal timeline: %w", err)
		}
	}
	if len(weightProfileRaw) > 0 && string(weightProfileRaw) != "{}" {
		if err := json.Unmarshal(weightProfileRaw, &m.WeightProfile); err != nil {
			return nil, fmt.Errorf("unmarshal weight_profile: %w", err)
		}
	}

	// Ensure slice fields are never nil so they serialize as [] not null.
	if m.Constraints == nil {
		m.Constraints = []Constraint{}
	}
	if m.CostCategories == nil {
		m.CostCategories = []string{}
	}
	if m.Timeline == nil {
		m.Timeline = []TimelineEvent{}
	}

	return &m, nil
}
