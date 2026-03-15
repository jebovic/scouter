import { useQuery } from '@tanstack/react-query'
import { getVolatilityCalendar } from '../api/volatilitycalendar'
import type { VolatilityCalendar } from '../api/volatilitycalendar'

export function useVolatilityCalendar(
  missionId: string,
  itemId: string,
): { calendar: VolatilityCalendar | null; isLoading: boolean } {
  const { data, isLoading } = useQuery({
    queryKey: ['volatility-calendar', missionId, itemId],
    queryFn: () => getVolatilityCalendar(missionId, itemId),
    staleTime: 2 * 60 * 60 * 1000, // 2 hours — matches backend cache TTL
    enabled: Boolean(missionId && itemId),
  })

  return { calendar: data ?? null, isLoading }
}
