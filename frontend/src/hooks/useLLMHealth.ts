import { useQuery } from '@tanstack/react-query'
import { getLLMHealth } from '../api/health'
import type { PoolHealth } from '../api/health'

export type LLMOverallStatus = 'healthy' | 'degraded' | 'down' | 'loading' | 'error'

export function deriveOverallStatus(
  data: PoolHealth | undefined,
  isLoading: boolean,
  isError: boolean,
): LLMOverallStatus {
  if (isLoading) return 'loading'
  if (isError || !data) return 'error'
  const models = data.models
  if (models.length === 0) return 'error'
  const healthyCount = models.filter(m => m.healthy).length
  if (healthyCount === models.length) return 'healthy'
  if (healthyCount === 0) return 'down'
  return 'degraded'
}

export function useLLMHealth() {
  const query = useQuery({
    queryKey: ['llm-health'],
    queryFn: getLLMHealth,
    refetchInterval: 60_000,
    staleTime: 55_000,
    retry: 2,
  })

  return {
    ...query,
    overallStatus: deriveOverallStatus(query.data, query.isLoading, query.isError),
  }
}
