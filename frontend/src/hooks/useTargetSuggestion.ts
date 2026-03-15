import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  applyTargetSuggestion,
  getTargetSuggestion,
  type TargetSuggestion,
} from '../api/targetsuggestion'

export function useTargetSuggestion(
  missionId: string,
  itemId: string,
): { suggestion: TargetSuggestion | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['target-suggestion', missionId, itemId],
    queryFn: () => getTargetSuggestion(missionId, itemId),
    staleTime: 2 * 60 * 60 * 1000, // 2 hours
    enabled: Boolean(missionId && itemId),
  })

  return { suggestion: data ?? null, isLoading }
}

export function useApplyTargetSuggestion(missionId: string, itemId: string) {
  const queryClient = useQueryClient()

  const { mutate, isPending } = useMutation({
    mutationFn: (target: 'conservative' | 'moderate' | 'aggressive') =>
      applyTargetSuggestion(missionId, itemId, target),
    onSuccess: () => {
      // Invalidate shopping item queries so target_price refreshes in UI.
      queryClient.invalidateQueries({ queryKey: ['shopping', missionId] })
      queryClient.invalidateQueries({ queryKey: ['target-suggestion', missionId, itemId] })
    },
  })

  return { apply: mutate, isApplying: isPending }
}
