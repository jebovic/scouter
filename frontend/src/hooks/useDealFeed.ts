import { useQuery } from '@tanstack/react-query'
import { getPromoFeed } from '../api/dealfeed'
import type { PromoFeedResult } from '../api/dealfeed'

/**
 * usePromoFeed fetches AI-generated French promo offers for a given query.
 * The query is only enabled when it has at least 2 characters.
 * Results are cached for 30 minutes (staleTime matches backend TTL).
 */
export function usePromoFeed(
  query: string,
  category?: string,
  enabled = true,
) {
  const isEnabled = enabled && query.length >= 2

  const { data, isLoading, isFetching, error } = useQuery<PromoFeedResult, Error>({
    queryKey: ['promo-feed', query, category ?? ''],
    queryFn: () => getPromoFeed(query, category),
    enabled: isEnabled,
    staleTime: 30 * 60 * 1000, // 30 min — matches backend cache TTL
  })

  return {
    result: data ?? null,
    offers: data?.offers ?? [],
    isLoading: isEnabled && (isLoading || isFetching),
    error,
  }
}
