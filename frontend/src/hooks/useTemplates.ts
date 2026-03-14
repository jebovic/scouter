import { useQuery } from '@tanstack/react-query'
import { listTemplates } from '../api'

export function useTemplates() {
  const { data: templates = [], isLoading, error } = useQuery({
    queryKey: ['templates'],
    queryFn: listTemplates,
    staleTime: 1000 * 60 * 60 * 24, // 24h — templates change only on backend deploy
  })
  return { templates, isLoading, error }
}
