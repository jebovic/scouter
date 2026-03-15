import { z } from 'zod'
import { apiFetch } from './client'

export const CategoryBreakdownSchema = z.object({
  category: z.string(),
  label: z.string(),
  totalSpent: z.number(),
  itemCount: z.number(),
  percentage: z.number(),
  icon: z.string(),
})

export const ExpenseSummarySchema = z.object({
  missionId: z.string().uuid(),
  totalItems: z.number(),
  totalSpent: z.number(),
  breakdowns: z.array(CategoryBreakdownSchema),
  topCategory: z.string(),
  insight: z.string(),
})

export type CategoryBreakdown = z.infer<typeof CategoryBreakdownSchema>
export type ExpenseSummary = z.infer<typeof ExpenseSummarySchema>

export async function getExpenseSummary(missionId: string): Promise<ExpenseSummary> {
  const raw = await apiFetch<unknown>(`/api/missions/${missionId}/expense-summary`)
  return ExpenseSummarySchema.parse(raw)
}
