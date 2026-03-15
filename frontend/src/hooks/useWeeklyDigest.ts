import { useQuery } from '@tanstack/react-query'
import { getWeeklyDigest } from '../api/weeklydigest'
import type { WeeklyDigest } from '../api/weeklydigest'

const ONE_HOUR = 60 * 60 * 1000

export function useWeeklyDigest() {
  const { data, isLoading, error } = useQuery<WeeklyDigest>({
    queryKey: ['weeklyDigest'],
    queryFn: getWeeklyDigest,
    staleTime: ONE_HOUR,
  })

  return { digest: data, isLoading, error }
}
