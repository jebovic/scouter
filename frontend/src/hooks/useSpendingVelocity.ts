import { useQuery } from '@tanstack/react-query'
import { getSpendingVelocity } from '../api/spendingvelocity'

export function useSpendingVelocity(missionId: string) {
  return useQuery({
    queryKey: ['spendingVelocity', missionId],
    queryFn: () => getSpendingVelocity(missionId),
    staleTime: 10 * 60 * 1000,
    enabled: !!missionId,
  })
}
