import { useQuery } from '@tanstack/react-query'
import { getSeasonalCalendar } from '../api/seasonalcalendar'
import type { CategoryCalendar } from '../api/seasonalcalendar'

export function useSeasonalCalendar(missionId: string) {
  const { data, isLoading, isError } = useQuery({
    queryKey: ['seasonalcalendar', missionId],
    queryFn: () => getSeasonalCalendar(missionId),
    staleTime: 24 * 60 * 60 * 1000,
    enabled: !!missionId,
  })
  return { data, isLoading, isError }
}
