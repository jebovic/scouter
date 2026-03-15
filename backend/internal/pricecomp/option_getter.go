package pricecomp

import (
	"context"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/option"
)

// OptionRepoGetter wraps an option.Repository to satisfy OptionNameGetter.
// This avoids a circular import: the pricecomp package stays free of pgx and
// the option domain model.
type OptionRepoGetter struct {
	repo option.Repository
}

// NewOptionRepoGetter returns an OptionNameGetter backed by an option.Repository.
func NewOptionRepoGetter(repo option.Repository) *OptionRepoGetter {
	return &OptionRepoGetter{repo: repo}
}

// GetOptionName resolves optionID (a UUID string) to the option's name.
// Returns ("", nil) when the option does not exist.
func (g *OptionRepoGetter) GetOptionName(ctx context.Context, optionID string) (string, error) {
	id, err := uuid.Parse(optionID)
	if err != nil {
		// Invalid UUID — treat as not found.
		return "", nil
	}
	opt, err := g.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if opt == nil {
		return "", nil
	}
	return opt.Name, nil
}
