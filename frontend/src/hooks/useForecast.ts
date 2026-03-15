import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchForecast, runForecast } from '../api/forecast'
import { useToast } from '../components/scouter'

export function useForecast(missionId: string) {
  const { data: forecast, isLoading, error } = useQuery({
    queryKey: ['forecast', missionId],
    queryFn: () => fetchForecast(missionId),
    staleTime: 5 * 60 * 1000,
    retry: false,
    enabled: !!missionId,
  })
  return { forecast, isLoading, error }
}

export function useRunForecast(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending, error } = useMutation({
    mutationFn: () => runForecast(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['forecast', missionId] })
      toast('Forecast complete', 'success')
    },
    onError: (err: unknown) =>
      toast(
        `Forecast failed: ${err instanceof Error ? err.message : 'Unknown error'}`,
        'error',
      ),
  })
  return { runForecast: mutateAsync, isPending, error }
}
