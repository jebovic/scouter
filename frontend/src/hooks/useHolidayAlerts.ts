import { useQuery } from '@tanstack/react-query'
import { getHolidaysAndEvents } from '../api/holidayalert'
import type { HolidayAlertResponse, HolidayEvent } from '../api/holidayalert'

export function useHolidayAlerts() {
  const { data, isLoading, error } = useQuery<HolidayAlertResponse, Error>({
    queryKey: ['holidays-and-events'],
    queryFn: getHolidaysAndEvents,
    staleTime: 24 * 60 * 60 * 1000, // 24 hours — static dataset
  })

  return {
    events: data?.events ?? ([] as HolidayEvent[]),
    nextShoppingEvent: data?.nextShoppingEvent ?? null,
    activeEvent: data?.activeEvent ?? null,
    isLoading,
    error,
  }
}
