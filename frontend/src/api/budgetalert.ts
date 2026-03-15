import { z } from 'zod'
import { apiFetch } from './client'

export const BudgetAlertSchema = z.object({
  missionId: z.string(),
  missionName: z.string(),
  totalBudget: z.number(),
  totalSpent: z.number(),
  percentage: z.number(),
  alertLevel: z.enum(['info', 'warning', 'critical']),
  message: z.string(),
  recommendations: z.array(z.string()),
})

export const BudgetAlertListSchema = z.object({
  alerts: z.array(BudgetAlertSchema),
  missionsChecked: z.number(),
})

export type BudgetAlert = z.infer<typeof BudgetAlertSchema>
export type BudgetAlertList = z.infer<typeof BudgetAlertListSchema>

export async function getBudgetAlerts(): Promise<BudgetAlertList> {
  const data = await apiFetch<unknown>('/api/budget-alerts')
  return BudgetAlertListSchema.parse(data)
}
