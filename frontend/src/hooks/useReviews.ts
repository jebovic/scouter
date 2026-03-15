import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getOptionReviews, refreshOptionReviews } from '../api/reviews'
import type { ReviewSummary } from '../api/reviews'

export function useOptionReviews(optionId: string) {
  const { data: summary, isLoading, error } = useQuery<ReviewSummary>({
    queryKey: ['reviews', optionId],
    queryFn: () => getOptionReviews(optionId),
    enabled: Boolean(optionId),
    staleTime: 6 * 60 * 60 * 1000, // 6h — matches backend cache
  })
  return { summary, isLoading, error }
}

export function useRefreshReviews(optionId: string) {
  const qc = useQueryClient()
  const { mutate: refresh, isPending } = useMutation({
    mutationFn: () => refreshOptionReviews(optionId),
    onSuccess: (data) => {
      qc.setQueryData(['reviews', optionId], data)
    },
  })
  return { refresh, isPending }
}
