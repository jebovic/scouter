import { z } from 'zod'
import { apiFetch } from './client'

export const ShoppingActionSchema = z.object({
  priority: z.number(),
  itemIds: z.array(z.string()),
  itemNames: z.array(z.string()),
  merchant: z.string(),
  action: z.string(),
  reason: z.string(),
  estSavings: z.number(),
  currency: z.string(),
})

export const OptimizedPlanSchema = z.object({
  missionId: z.string(),
  generatedAt: z.string(),
  totalItems: z.number(),
  actions: z.array(ShoppingActionSchema),
  summary: z.string(),
})

export type ShoppingAction = z.infer<typeof ShoppingActionSchema>
export type OptimizedPlan = z.infer<typeof OptimizedPlanSchema>

export async function optimizeShoppingList(missionSlug: string): Promise<OptimizedPlan> {
  const data = await apiFetch<unknown>(`/api/missions/${missionSlug}/optimize`, {
    method: 'POST',
  })
  return OptimizedPlanSchema.parse(data)
}
