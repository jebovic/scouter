import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { getAdvice } from '../api/purchaseadvisor'
import type { PurchaseAdvice } from '../api/purchaseadvisor'

/**
 * usePurchaseAdvice triggers the AI Purchase Advisor for a shopping item.
 *
 * Uses useMutation (not useQuery) because the endpoint is a POST and the
 * advisor is invoked on demand, not on mount. Results are cached in local
 * component state so re-renders don't re-fetch.
 */
export function usePurchaseAdvice(itemId: string) {
  const [cached, setCached] = useState<PurchaseAdvice | null>(null)

  const mutation = useMutation<PurchaseAdvice, Error>({
    mutationFn: () => getAdvice(itemId),
    onSuccess: (data) => setCached(data),
  })

  return {
    advice: cached,
    isPending: mutation.isPending,
    error: mutation.error,
    getAdvice: () => {
      if (cached) return
      mutation.mutate()
    },
  }
}
