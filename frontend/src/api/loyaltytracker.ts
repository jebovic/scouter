import { z } from 'zod'
import { apiFetch } from './client'

export const MerchantSummarySchema = z.object({
  merchant: z.string(),
  total: z.number(),
  program: z.string(),
  tier: z.string().optional(),
  estimatedCashback: z.number(),
  estimatedPoints: z.number(),
})

export const LoyaltySummarySchema = z.object({
  merchants: z.array(MerchantSummarySchema),
  totalEstimatedCashback: z.number(),
})

export type MerchantSummary = z.infer<typeof MerchantSummarySchema>
export type LoyaltySummary = z.infer<typeof LoyaltySummarySchema>

export async function getLoyaltySummary(missionId: string): Promise<LoyaltySummary> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/loyalty-summary`)
  return LoyaltySummarySchema.parse(data)
}
