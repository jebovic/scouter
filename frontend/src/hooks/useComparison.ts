import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchWeights, setWeight, deleteWeight, fetchMatrix } from '../api/comparison'
import type { Weight, MatrixResponse } from '../api/comparison'

const STALE_MS = 30_000

export function useWeights(missionId: string) {
  const { data, isLoading } = useQuery<Weight[]>({
    queryKey: ['comparison-weights', missionId],
    queryFn: () => fetchWeights(missionId),
    enabled: Boolean(missionId),
    staleTime: STALE_MS,
  })
  return { weights: data ?? [], isLoading }
}

export function useMatrix(missionId: string) {
  const { data, isLoading, refetch } = useQuery<MatrixResponse>({
    queryKey: ['matrix', missionId],
    queryFn: () => fetchMatrix(missionId),
    enabled: Boolean(missionId),
    staleTime: STALE_MS,
  })
  return { matrix: data ?? null, isLoading, refetch }
}

export function useSetWeight(missionId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ criteria, weight }: { criteria: string; weight: number }) =>
      setWeight(missionId, criteria, weight),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['comparison-weights', missionId] })
      void queryClient.invalidateQueries({ queryKey: ['matrix', missionId] })
    },
  })
}

export function useDeleteWeight(missionId: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (criteria: string) => deleteWeight(missionId, criteria),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['comparison-weights', missionId] })
      void queryClient.invalidateQueries({ queryKey: ['matrix', missionId] })
    },
  })
}
