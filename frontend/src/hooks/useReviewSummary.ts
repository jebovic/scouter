import { useQuery } from '@tanstack/react-query'
import { getReviewSummary } from '../api/reviewsummary'
import type { ItemReviewSummary } from '../api/reviewsummary'

const STALE_MS = 30 * 60 * 1000 // 30 min — matches backend cache TTL

export function useReviewSummary(itemId: string, itemName: string) {
  const { data: summary, isLoading, error } = useQuery<ItemReviewSummary>({
    queryKey: ['review-summary', itemId],
    queryFn: () => getReviewSummary(itemId, itemName),
    enabled: Boolean(itemId),
    staleTime: STALE_MS,
  })
  return { summary, isLoading, error }
}
