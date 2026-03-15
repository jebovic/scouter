import { useQuery } from '@tanstack/react-query'
import { listDealEvents } from '../api/dealCalendar'
import type { DealEvent } from '../api/dealCalendar'

export function useDealCalendar(params?: { category?: string; upcoming?: boolean }) {
  const { data, isLoading, error } = useQuery<DealEvent[], Error>({
    queryKey: ['deal-calendar', params?.category ?? '', params?.upcoming ?? false],
    queryFn: () => listDealEvents(params),
    staleTime: 24 * 60 * 60 * 1000,
  })

  return {
    events: data ?? [],
    isLoading,
    error,
  }
}
