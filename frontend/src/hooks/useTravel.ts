import { useQuery } from '@tanstack/react-query'
import { searchFlights, searchTrains } from '../api/travel'
import type { FlightOffer, TrainOffer } from '../api/travel'

export function useFlightSearch(from: string, to: string, date: string) {
  const enabled = Boolean(from && to && date)
  const { data: flights = [], isLoading, error } = useQuery<FlightOffer[]>({
    queryKey: ['travel', 'flights', from, to, date],
    queryFn: () => searchFlights(from, to, date),
    enabled,
    staleTime: 60 * 60 * 1000, // 1h — matches backend cache
  })
  return { flights, isLoading, error }
}

export function useTrainSearch(from: string, to: string, date: string) {
  const enabled = Boolean(from && to && date)
  const { data: trains = [], isLoading, error } = useQuery<TrainOffer[]>({
    queryKey: ['travel', 'trains', from, to, date],
    queryFn: () => searchTrains(from, to, date),
    enabled,
    staleTime: 60 * 60 * 1000,
  })
  return { trains, isLoading, error }
}
