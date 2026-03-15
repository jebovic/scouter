import { useMutation, useQueryClient } from '@tanstack/react-query'
import { optimizeShopping } from '../api/shoppingoptimizer'

export function useShoppingOptimizer(missionId: string) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => optimizeShopping(missionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shopping', missionId] })
    },
  })
}
