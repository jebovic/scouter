import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchWishList, createWishListItem, deleteWishListItem } from '../api/wishlist'
import type { WishListItem } from '../api/wishlist'

const QUERY_KEY = ['wishlist'] as const

export function useWishList() {
  const { data, isLoading } = useQuery<WishListItem[]>({
    queryKey: QUERY_KEY,
    queryFn: fetchWishList,
    staleTime: 60_000,
  })

  return { items: data ?? [], isLoading }
}

export function useCreateWishListItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: createWishListItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })
}

export function useDeleteWishListItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: deleteWishListItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: QUERY_KEY })
    },
  })
}
