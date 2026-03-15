import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getWatchlist, addToWatchlist, removeFromWatchlist } from '../api/watchlist'

export function useWatchlist() {
  const qc = useQueryClient()
  const query = useQuery({
    queryKey: ['watchlist'],
    queryFn: getWatchlist,
    staleTime: 30_000,
  })

  const add = useMutation({
    mutationFn: addToWatchlist,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['watchlist'] }),
  })

  const remove = useMutation({
    mutationFn: removeFromWatchlist,
    onSuccess: () => qc.invalidateQueries({ queryKey: ['watchlist'] }),
  })

  return { ...query, add, remove }
}
