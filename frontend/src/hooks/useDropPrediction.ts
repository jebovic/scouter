import { useQuery } from '@tanstack/react-query'
import { getDropPrediction } from '../api/dropreporter'
import type { DropPrediction } from '../api/dropreporter'

export function useDropPrediction(
  missionId: string,
  itemId: string,
): { prediction: DropPrediction | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['drop-prediction', missionId, itemId],
    queryFn: () => getDropPrediction(missionId, itemId),
    staleTime: 60 * 60 * 1000, // 60 minutes
    enabled: Boolean(missionId && itemId),
  })

  return { prediction: data ?? null, isLoading }
}
