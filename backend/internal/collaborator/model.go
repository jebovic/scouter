package collaborator

import "time"

// Role represents a collaborator's permission level on a mission.
type Role string

const (
	RoleViewer Role = "viewer"
	RoleEditor Role = "editor"
	RoleVoter  Role = "voter"
)

// Valid returns true when the role is one of the three accepted values.
func (r Role) Valid() bool {
	return r == RoleViewer || r == RoleEditor || r == RoleVoter
}

// Collaborator is a user who has been invited to collaborate on a mission.
type Collaborator struct {
	ID          string     `json:"id"`
	MissionID   string     `json:"missionId"`
	InviteToken string     `json:"inviteToken"`
	Email       string     `json:"email"`
	DisplayName *string    `json:"displayName"`
	Role        Role       `json:"role"`
	JoinedAt    *time.Time `json:"joinedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// CreateInviteRequest is the payload for inviting a collaborator.
type CreateInviteRequest struct {
	Email string `json:"email"`
	Role  Role   `json:"role"`
}

// JoinRequest is the payload a collaborator submits when accepting an invite.
type JoinRequest struct {
	DisplayName string `json:"displayName"`
}
