import { z } from 'zod'
import { apiFetch } from './client'

const AllocationItemSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  allocatedBudget: z.number(),
  priority: z.number(),
  reason: z.string(),
  feasible: z.boolean(),
})

export const BudgetAdviceSchema = z.object({
  missionId: z.string(),
  totalBudget: z.number(),
  spent: z.number(),
  remaining: z.number(),
  allocations: z.array(AllocationItemSchema),
  totalAllocated: z.number(),
  surplus: z.number(),
  advice: z.string(),
})

export type AllocationItem = z.infer<typeof AllocationItemSchema>
export type BudgetAdvice = z.infer<typeof BudgetAdviceSchema>

export async function getBudgetAdvice(missionId: string): Promise<BudgetAdvice> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/budget-advice`)
  return BudgetAdviceSchema.parse(data)
}
