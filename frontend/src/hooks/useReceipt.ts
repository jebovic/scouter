import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { analyzeReceipt, listReceipts } from '../api/receipt'

export function useReceipts(slug: string) {
  return useQuery({
    queryKey: ['receipts', slug],
    queryFn: () => listReceipts(slug),
    staleTime: 2 * 60 * 1000, // 2 minutes
    enabled: Boolean(slug),
  })
}

export function useAnalyzeReceipt(slug: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ rawText, currency }: { rawText: string; currency?: string }) =>
      analyzeReceipt(slug, rawText, currency),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['receipts', slug] })
    },
  })
}
