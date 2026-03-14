import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listShoppingItems,
  createShoppingItem,
  updateShoppingItem,
  deleteShoppingItem,
  pinShoppingItem,
  deletePinnedShoppingItems,
  addPriceSnapshot,
  listPriceSnapshots,
} from '../api'
import { useToast } from '../components/scouter'
import type { ShoppingItemCreateRequest, ShoppingItemUpdateRequest, PriceSnapshotRequest } from '../types'

export function useShopping(missionId: string) {
  const {
    data: items = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['shopping', missionId],
    queryFn: () => listShoppingItems(missionId),
    enabled: Boolean(missionId),
  })
  return { items, isLoading, error }
}

export function useCreateShoppingItem(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (req: ShoppingItemCreateRequest) => createShoppingItem(missionId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Item added', 'success')
    },
    onError: (err: unknown) => toast(`Failed to add item: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { createItem: mutateAsync, isPending }
}

export function useUpdateShoppingItem(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: ({ itemId, req }: { itemId: string; req: ShoppingItemUpdateRequest }) =>
      updateShoppingItem(missionId, itemId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Item updated', 'success')
    },
    onError: (err: unknown) => toast(`Failed to update item: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { updateItem: mutateAsync, isPending }
}

export function useDeleteShoppingItem(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (itemId: string) => deleteShoppingItem(missionId, itemId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Item removed', 'success')
    },
    onError: (err: unknown) => toast(`Failed to remove item: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { deleteItem: mutateAsync, isPending }
}

export function usePinShoppingItem(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (itemId: string) => pinShoppingItem(missionId, itemId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Item pinned', 'success')
    },
    onError: (err: unknown) => toast(`Failed to pin item: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { pinItem: mutateAsync, isPending }
}

export function useDeletePinnedShoppingItems(missionId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: () => deletePinnedShoppingItems(missionId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Pinned items cleared', 'success')
    },
    onError: (err: unknown) => toast(`Failed to clear pinned items: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { deletePinnedItems: mutateAsync, isPending }
}

export function usePriceSnapshots(missionId: string, itemId: string) {
  const {
    data: snapshots = [],
    isLoading,
  } = useQuery({
    queryKey: ['snapshots', missionId, itemId],
    queryFn: () => listPriceSnapshots(missionId, itemId),
    enabled: Boolean(missionId) && Boolean(itemId),
  })
  return { snapshots, isLoading }
}

export function useAddPriceSnapshot(missionId: string, itemId: string) {
  const qc = useQueryClient()
  const { toast } = useToast()
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (req: PriceSnapshotRequest) => addPriceSnapshot(missionId, itemId, req),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['snapshots', missionId, itemId] })
      qc.invalidateQueries({ queryKey: ['shopping', missionId] })
      toast('Price snapshot added', 'success')
    },
    onError: (err: unknown) => toast(`Failed to add price snapshot: ${err instanceof Error ? err.message : 'Unknown error'}`, 'error'),
  })
  return { addSnapshot: mutateAsync, isPending }
}
