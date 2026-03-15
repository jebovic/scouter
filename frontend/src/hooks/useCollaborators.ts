import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listCollaborators,
  createInvite,
  getInvite,
  joinMission,
  removeCollaborator,
} from '../api/collaborator'
import type { CreateInviteResponse } from '../api/collaborator'

export function useCollaborators(missionId: string) {
  return useQuery({
    queryKey: ['collaborators', missionId],
    queryFn: () => listCollaborators(missionId),
    enabled: !!missionId,
    staleTime: 30_000,
  })
}

export function useCreateInvite(missionId: string) {
  const qc = useQueryClient()
  return useMutation<CreateInviteResponse, Error, { email: string; role: string }>({
    mutationFn: ({ email, role }) => createInvite(missionId, email, role),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['collaborators', missionId] }),
  })
}

export function useRemoveCollaborator(missionId: string) {
  const qc = useQueryClient()
  return useMutation<void, Error, string>({
    mutationFn: (collaboratorId: string) => removeCollaborator(missionId, collaboratorId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['collaborators', missionId] }),
  })
}

export function useGetInvite(token: string) {
  return useQuery({
    queryKey: ['invite', token],
    queryFn: () => getInvite(token),
    enabled: !!token,
    retry: false,
  })
}

export function useJoinMission(token: string) {
  return useMutation({
    mutationFn: (displayName: string) => joinMission(token, displayName),
  })
}
