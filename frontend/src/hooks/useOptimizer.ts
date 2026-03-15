import { useMutation } from '@tanstack/react-query'
import { optimizeShoppingList } from '../api/optimizer'

export function useOptimizeShoppingList(missionSlug: string) {
  return useMutation({
    mutationFn: () => optimizeShoppingList(missionSlug),
  })
}
