package collaborator_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jibei/scouter/internal/collaborator"
)

func TestRole_Valid(t *testing.T) {
	assert.True(t, collaborator.Role("viewer").Valid())
	assert.True(t, collaborator.Role("editor").Valid())
	assert.True(t, collaborator.Role("voter").Valid())
	assert.False(t, collaborator.Role("admin").Valid())
	assert.False(t, collaborator.Role("").Valid())
	assert.False(t, collaborator.Role("superuser").Valid())
}

func TestCollaborator_JoinedAt_InitiallyNil(t *testing.T) {
	c := &collaborator.Collaborator{}
	assert.Nil(t, c.JoinedAt, "new collaborator should not have a joined_at timestamp")
}

func TestCollaborator_JoinedAt_AfterJoin(t *testing.T) {
	c := &collaborator.Collaborator{}
	now := time.Now()
	c.JoinedAt = &now
	assert.NotNil(t, c.JoinedAt)
	assert.WithinDuration(t, now, *c.JoinedAt, time.Second)
}

func TestCreateInviteRequest_Fields(t *testing.T) {
	req := collaborator.CreateInviteRequest{
		Email: "alice@example.com",
		Role:  collaborator.RoleVoter,
	}
	assert.Equal(t, "alice@example.com", req.Email)
	assert.True(t, req.Role.Valid())
}

func TestJoinRequest_DisplayName(t *testing.T) {
	req := collaborator.JoinRequest{DisplayName: "Alice"}
	assert.Equal(t, "Alice", req.DisplayName)
}

func TestRole_Constants(t *testing.T) {
	assert.Equal(t, collaborator.Role("viewer"), collaborator.RoleViewer)
	assert.Equal(t, collaborator.Role("editor"), collaborator.RoleEditor)
	assert.Equal(t, collaborator.Role("voter"), collaborator.RoleVoter)
}
