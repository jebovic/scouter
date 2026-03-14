import { useQuery } from '@tanstack/react-query'
import { listAgentRuns } from '../api'

export function useAgentRuns(missionId: string, agentType?: 'research' | 'pricing') {
  const {
    data: runs = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['agent-runs', missionId, agentType],
    queryFn: () => listAgentRuns(missionId, agentType),
    enabled: Boolean(missionId),
  })
  return { runs, isLoading, error }
}
