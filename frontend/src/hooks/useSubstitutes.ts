import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchSubstitutes } from '../api/substitutes'
import type { SubstituteResponse } from '../api/substitutes'

/** Reads the substitute suggestions from cache (2h stale). Does not auto-fetch. */
export function useSubstitutes(optionId: string | undefined) {
  return useQuery<SubstituteResponse>({
    queryKey: ['substitutes', optionId],
    queryFn: () => fetchSubstitutes(optionId!),
    enabled: Boolean(optionId),
    staleTime: 2 * 60 * 60 * 1000, // 2h — matches backend cache TTL
  })
}

/** Imperative mutation to force-fetch (or refresh) substitutes for an option. */
export function useFetchSubstitutes(optionId: string | undefined) {
  const qc = useQueryClient()
  const { mutate: fetch, isPending } = useMutation({
    mutationFn: () => {
      if (!optionId) return Promise.reject(new Error('optionId is required'))
      return fetchSubstitutes(optionId)
    },
    onSuccess: (data) => {
      qc.setQueryData(['substitutes', optionId], data)
    },
  })
  return { fetch, isPending }
}
