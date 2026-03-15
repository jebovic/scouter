import { useQuery } from '@tanstack/react-query'
import { getElasticity } from '../api/elasticity'
import type { ElasticityResult } from '../api/elasticity'

export function useElasticity(
  missionId: string,
  itemId: string,
): { result: ElasticityResult | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['elasticity', missionId, itemId],
    queryFn: () => getElasticity(missionId, itemId),
    staleTime: 60 * 60 * 1000, // 1 hour — matches backend cache TTL
    enabled: Boolean(missionId && itemId),
  })

  return { result: data ?? null, isLoading }
}
