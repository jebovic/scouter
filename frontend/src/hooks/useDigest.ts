import { useQuery } from '@tanstack/react-query'
import { fetchWeeklyDigest } from '../api/digest'
import type { WeeklyDigest } from '../api/digest'

/** Fetches the weekly price drop digest. 30-minute stale time. */
export function useWeeklyDigest(days = 7) {
  const { data, isLoading, error } = useQuery<WeeklyDigest>({
    queryKey: ['weeklyDigest', days],
    queryFn: () => fetchWeeklyDigest(days),
    staleTime: 30 * 60 * 1000, // 30 min
  })

  return { digest: data, isLoading, error }
}
