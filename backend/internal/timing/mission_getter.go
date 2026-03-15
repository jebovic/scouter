package timing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MissionInfoGetter is the narrow interface used by Handler to resolve a
// mission ID to its metadata. Tests can inject a fake.
type MissionInfoGetter interface {
	// GetMissionInfo returns the mission name, category, budget in EUR, and
	// creation time for the given missionID. Returns an error when the mission
	// does not exist or the lookup fails.
	GetMissionInfo(ctx context.Context, missionID string) (name, category string, budgetEur float64, createdAt time.Time, err error)
}

// MissionRepoGetter implements MissionInfoGetter using a pgxpool.Pool.
type MissionRepoGetter struct {
	pool *pgxpool.Pool
}

// NewMissionRepoGetter returns a MissionInfoGetter backed by the given pool.
func NewMissionRepoGetter(pool *pgxpool.Pool) *MissionRepoGetter {
	return &MissionRepoGetter{pool: pool}
}

// GetMissionInfo queries the missions table for the given missionID.
// It returns a non-nil error when the mission is not found or the query fails.
func (g *MissionRepoGetter) GetMissionInfo(ctx context.Context, missionID string) (name, category string, budgetEur float64, createdAt time.Time, err error) {
	id, parseErr := uuid.Parse(missionID)
	if parseErr != nil {
		err = fmt.Errorf("invalid mission id %q: %w", missionID, parseErr)
		return
	}

	row := g.pool.QueryRow(ctx,
		`SELECT name, category, budget, created_at FROM missions WHERE id = $1`,
		id,
	)
	if scanErr := row.Scan(&name, &category, &budgetEur, &createdAt); scanErr != nil {
		err = fmt.Errorf("mission not found: %w", scanErr)
		return
	}
	return
}
