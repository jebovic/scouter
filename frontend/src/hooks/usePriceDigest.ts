import { useQuery } from '@tanstack/react-query'
import { getPriceDigest } from '../api/pricedigest'
import type { Digest } from '../api/pricedigest'

const TEN_MINUTES = 10 * 60 * 1000

/** Fetches the 24-hour price change digest. 10-minute stale time, refetch every 10 min. */
export function usePriceDigest() {
  const { data, isLoading, error } = useQuery<Digest>({
    queryKey: ['priceDigest'],
    queryFn: getPriceDigest,
    staleTime: TEN_MINUTES,
    refetchInterval: TEN_MINUTES,
  })

  return { digest: data, isLoading, error }
}
