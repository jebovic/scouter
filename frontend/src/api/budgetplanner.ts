import { z } from 'zod'
import { apiFetch } from './client'

export const PlannedItemSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  price: z.number(),
  status: z.string(),
  priority: z.number(),
  reason: z.string(),
  cumulativeCost: z.number(),
})

export const PurchasePlanSchema = z.object({
  missionId: z.string(),
  totalBudget: z.number(),
  totalEstimate: z.number(),
  budgetStatus: z.enum(['under', 'over', 'tight']),
  sequence: z.array(PlannedItemSchema),
  deferred: z.array(PlannedItemSchema),
  savings: z.number(),
  generatedAt: z.string(),
})

export type PlannedItem = z.infer<typeof PlannedItemSchema>
export type PurchasePlan = z.infer<typeof PurchasePlanSchema>

export async function getBudgetPlan(missionId: string): Promise<PurchasePlan> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/budget-plan`)
  return PurchasePlanSchema.parse(data)
}
