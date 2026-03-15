import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listMissions,
  getMission,
  createMission,
  updateMission,
  deleteMission,
  duplicateMission,
} from '../api'
import { useToast } from '../components/scouter'
import type { MissionCreateRequest, MissionUpdateRequest } from '../types'

export function useMissions() {
  const {
    data: missions = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['missions'],
    queryFn: listMissions,
  })
  return { missions, isLoading, error }
}

export function useMission(slug: string) {
  const {
    data: mission,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['missions', slug],
    queryFn: () => getMission(slug),
    enabled: Boolean(slug),
  })
  return { mission, isLoading, error }
}

export function useCreateMission() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (req: MissionCreateRequest) => createMission(req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['missions'] })
      toast('Mission created', 'success')
    },
    onError: (err: unknown) => toast(`Failed to create mission: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { createMission: mutateAsync, isPending }
}

export function useUpdateMission(slug: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (req: MissionUpdateRequest) => updateMission(slug, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['missions'] })
      qc.invalidateQueries({ queryKey: ['missions', slug] })
      toast('Mission updated', 'success')
    },
    onError: (err: unknown) => toast(`Failed to update mission: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { updateMission: mutateAsync, isPending }
}

export function useDeleteMission() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (slug: string) => deleteMission(slug),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['missions'] })
      toast('Mission deleted', 'success')
    },
    onError: (err: unknown) => toast(`Failed to delete mission: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { deleteMission: mutateAsync, isPending }
}

export function useDuplicateMission() {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (slug: string) => duplicateMission(slug),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['missions'] })
      toast('Mission duplicated', 'success')
    },
    onError: (err: unknown) => toast(`Failed to duplicate mission: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { duplicateMission: mutateAsync, isPending }
}
