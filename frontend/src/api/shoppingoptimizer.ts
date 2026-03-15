import { z } from 'zod'
import { apiFetch } from './client'

export const OptimizedItemSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  merchant: z.string(),
  price: z.number(),
  priorityScore: z.number(),
  reasons: z.array(z.string()),
  urgencyLevel: z.enum(['now', 'soon', 'wait']),
  savingsPotential: z.number(),
})

export const OptimizeResponseSchema = z.object({
  items: z.array(OptimizedItemSchema),
  totalBudget: z.number(),
  totalOptimized: z.number(),
  summary: z.string(),
})

export type OptimizedItem = z.infer<typeof OptimizedItemSchema>
export type OptimizeResponse = z.infer<typeof OptimizeResponseSchema>

export async function optimizeShopping(missionId: string): Promise<OptimizeResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/shopping/optimize`, {
    method: 'POST',
  })
  return OptimizeResponseSchema.parse(data)
}
