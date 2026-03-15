import { z } from 'zod'

export const CollaboratorSchema = z.object({
  id: z.string(),
  missionId: z.string(),
  inviteToken: z.string(),
  email: z.string(),
  displayName: z.string().nullable(),
  role: z.enum(['viewer', 'editor', 'voter']),
  joinedAt: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Collaborator = z.infer<typeof CollaboratorSchema>

export const CollaboratorsListSchema = z.array(CollaboratorSchema)

export const CreateInviteResponseSchema = z.object({
  collaborator: CollaboratorSchema,
  inviteUrl: z.string(),
})

export type CreateInviteResponse = z.infer<typeof CreateInviteResponseSchema>

const API_BASE = import.meta.env.VITE_API_URL ?? '/api'

export async function createInvite(
  missionId: string,
  email: string,
  role: string,
): Promise<CreateInviteResponse> {
  const res = await fetch(`${API_BASE}/missions/${missionId}/invites`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, role }),
  })
  if (!res.ok) throw new Error('Failed to create invite')
  return CreateInviteResponseSchema.parse(await res.json())
}

export async function getInvite(token: string): Promise<Collaborator> {
  const res = await fetch(`${API_BASE}/invites/${token}`)
  if (!res.ok) throw new Error('Invalid invite link')
  return CollaboratorSchema.parse(await res.json())
}

export async function joinMission(token: string, displayName: string): Promise<Collaborator> {
  const res = await fetch(`${API_BASE}/invites/${token}/join`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ displayName }),
  })
  if (!res.ok) throw new Error('Failed to join mission')
  return CollaboratorSchema.parse(await res.json())
}

export async function listCollaborators(missionId: string): Promise<Collaborator[]> {
  const res = await fetch(`${API_BASE}/missions/${missionId}/collaborators`)
  if (!res.ok) throw new Error('Failed to load collaborators')
  return CollaboratorsListSchema.parse(await res.json())
}

export async function removeCollaborator(
  missionId: string,
  collaboratorId: string,
): Promise<void> {
  const res = await fetch(
    `${API_BASE}/missions/${missionId}/collaborators/${collaboratorId}`,
    { method: 'DELETE' },
  )
  if (!res.ok) throw new Error('Failed to remove collaborator')
}
