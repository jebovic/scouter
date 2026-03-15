package substitute

import (
	"context"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/option"
)

// OptionRepoGetter wraps an option.Repository to satisfy OptionInfoGetter.
// This keeps the substitute package free of direct pgx dependencies.
type OptionRepoGetter struct {
	repo option.Repository
}

// NewOptionRepoGetter returns an OptionInfoGetter backed by an option.Repository.
func NewOptionRepoGetter(repo option.Repository) *OptionRepoGetter {
	return &OptionRepoGetter{repo: repo}
}

// GetOptionInfo resolves optionID (a UUID string) to the option's name,
// category, and best price. Returns ("", "", 0, nil) when the option does
// not exist.
func (g *OptionRepoGetter) GetOptionInfo(ctx context.Context, optionID string) (name, category string, budget float64, err error) {
	id, parseErr := uuid.Parse(optionID)
	if parseErr != nil {
		// Invalid UUID — treat as not found.
		return "", "", 0, nil
	}
	opt, err := g.repo.GetByID(ctx, id)
	if err != nil {
		return "", "", 0, err
	}
	if opt == nil {
		return "", "", 0, nil
	}
	if opt.PriceRange != nil {
		budget = opt.PriceRange.Best
	}
	return opt.Name, opt.Category, budget, nil
}
