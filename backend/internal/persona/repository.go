package persona

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines data access for persona records.
type Repository interface {
	Save(ctx context.Context, p Persona) (Persona, error)
	GetLatest(ctx context.Context) (Persona, error)
}

type pgRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a PostgreSQL-backed persona repository.
func NewRepository(pool *pgxpool.Pool) Repository {
	return &pgRepository{pool: pool}
}

func (r *pgRepository) Save(ctx context.Context, p Persona) (Persona, error) {
	traitsJSON, err := json.Marshal(p.Traits)
	if err != nil {
		return Persona{}, fmt.Errorf("marshal traits: %w", err)
	}
	tipsJSON, err := json.Marshal(p.Tips)
	if err != nil {
		return Persona{}, fmt.Errorf("marshal tips: %w", err)
	}

	row := r.pool.QueryRow(ctx, `
		INSERT INTO persona (archetype, traits, tips, summary)
		VALUES ($1, $2::jsonb, $3::jsonb, $4)
		RETURNING id, archetype, traits, tips, summary, created_at`,
		p.Archetype,
		traitsJSON,
		tipsJSON,
		p.Summary,
	)

	return scanPersona(row)
}

func (r *pgRepository) GetLatest(ctx context.Context) (Persona, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, archetype, traits, tips, summary, created_at
		FROM persona
		ORDER BY created_at DESC
		LIMIT 1`)

	p, err := scanPersona(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Persona{}, ErrNotFound
	}
	return p, err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPersona(s rowScanner) (Persona, error) {
	var p Persona
	var traitsRaw, tipsRaw []byte

	err := s.Scan(
		&p.ID,
		&p.Archetype,
		&traitsRaw,
		&tipsRaw,
		&p.Summary,
		&p.CreatedAt,
	)
	if err != nil {
		return Persona{}, err
	}

	if len(traitsRaw) > 0 {
		if err := json.Unmarshal(traitsRaw, &p.Traits); err != nil {
			return Persona{}, fmt.Errorf("unmarshal traits: %w", err)
		}
	}
	if p.Traits == nil {
		p.Traits = []string{}
	}

	if len(tipsRaw) > 0 {
		if err := json.Unmarshal(tipsRaw, &p.Tips); err != nil {
			return Persona{}, fmt.Errorf("unmarshal tips: %w", err)
		}
	}
	if p.Tips == nil {
		p.Tips = []string{}
	}

	return p, nil
}
