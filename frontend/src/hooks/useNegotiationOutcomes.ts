import { useQuery } from '@tanstack/react-query'
import { getNegotiationOutcomes } from '../api/negotiationoutcome'

export function useNegotiationOutcomes(missionId: string) {
  const { data, isLoading } = useQuery({
    queryKey: ['negotiation-outcomes', missionId],
    queryFn: () => getNegotiationOutcomes(missionId),
    staleTime: 1000 * 60 * 30,
    enabled: !!missionId,
  })
  return { summary: data, isLoading }
}
