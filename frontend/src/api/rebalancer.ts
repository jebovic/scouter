import { z } from 'zod'
import { apiFetch } from './client'

export const EnvelopeAdjustmentSchema = z.object({
  envelopeId: z.string(),
  envelopeName: z.string(),
  currentMonthly: z.number(),
  suggestedMonthly: z.number(),
  deltaPercent: z.number(),
  reason: z.string(),
})

export const RebalancePlanSchema = z.object({
  adjustments: z.array(EnvelopeAdjustmentSchema),
  summary: z.string(),
  totalCurrent: z.number(),
  totalSuggested: z.number(),
  cachedAt: z.number(),
})

export type EnvelopeAdjustment = z.infer<typeof EnvelopeAdjustmentSchema>
export type RebalancePlan = z.infer<typeof RebalancePlanSchema>

export async function getRebalancePlan(): Promise<RebalancePlan> {
  const data = await apiFetch<unknown>('/api/budget/rebalance')
  return RebalancePlanSchema.parse(data)
}
