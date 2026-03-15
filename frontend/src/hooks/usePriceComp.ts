import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getPriceComparison, refreshPriceComparison } from '../api/pricecomp'
import type { PriceComparison } from '../api/pricecomp'

export function usePriceComparison(optionId: string) {
  const { data: comparison, isLoading, error } = useQuery<PriceComparison>({
    queryKey: ['pricecomp', optionId],
    queryFn: () => getPriceComparison(optionId),
    enabled: Boolean(optionId),
    staleTime: 30 * 60 * 1000, // 30min — matches backend cache
  })
  return { comparison, isLoading, error }
}

export function useRefreshPriceComparison(optionId: string) {
  const qc = useQueryClient()
  const { mutate: refresh, isPending } = useMutation({
    mutationFn: () => refreshPriceComparison(optionId),
    onSuccess: (data) => {
      qc.setQueryData(['pricecomp', optionId], data)
    },
  })
  return { refresh, isPending }
}
