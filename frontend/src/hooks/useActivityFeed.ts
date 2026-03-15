import { useQuery } from '@tanstack/react-query'
import { getActivityFeed } from '../api/activityfeed'
import type { ActivityFeed } from '../api/activityfeed'

const FIVE_MINUTES = 5 * 60 * 1000
const ONE_MINUTE = 60 * 1000

export function useActivityFeed() {
  const { data, isLoading, error } = useQuery<ActivityFeed>({
    queryKey: ['activityFeed'],
    queryFn: getActivityFeed,
    staleTime: FIVE_MINUTES,
    refetchInterval: ONE_MINUTE,
  })

  return { feed: data, isLoading, error }
}
