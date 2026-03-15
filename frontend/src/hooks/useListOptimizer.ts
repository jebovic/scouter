import { useQuery } from '@tanstack/react-query'
import { getOptimizedList } from '../api/listoptimizer'

export function useListOptimizer(missionId: string) {
  const { data, isLoading, error } = useQuery({
    queryKey: ['listoptimizer', missionId],
    queryFn: () => getOptimizedList(missionId),
    staleTime: 15 * 60 * 1000,
    enabled: !!missionId,
  })
  return { data, isLoading, error }
}
