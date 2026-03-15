package collaborator_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jibei/scouter/internal/collaborator"
)

func TestService_CreateInvite_InvalidRole(t *testing.T) {
	// Service rejects invalid roles before hitting the repository.
	svc := collaborator.NewService(nil) // repo won't be called
	_, err := svc.CreateInvite(nil, "mission-1", "bob@example.com", collaborator.Role("admin")) //nolint:staticcheck
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestService_JoinMission_EmptyDisplayName(t *testing.T) {
	svc := collaborator.NewService(nil)
	_, err := svc.JoinMission(nil, "some-token", "") //nolint:staticcheck
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "display name is required")
}
