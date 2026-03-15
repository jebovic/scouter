import { useQuery } from '@tanstack/react-query'
import { getSaleCalendar } from '../api/salecalendar'
import type { SaleCalendar } from '../api/salecalendar'

export function useSaleCalendar() {
  const { data, isLoading, error } = useQuery<SaleCalendar, Error>({
    queryKey: ['sale-calendar'],
    queryFn: getSaleCalendar,
    staleTime: 60 * 60 * 1000, // 1 hour
  })

  return {
    calendar: data ?? null,
    events: data?.events ?? [],
    nextEvent: data?.nextEvent ?? null,
    activeEvent: data?.activeEvent ?? null,
    isLoading,
    error,
  }
}
