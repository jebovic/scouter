import { z } from 'zod'
import { apiFetch } from './client'

const MerchantCashbackSchema = z.object({
  merchant: z.string(),
  platform: z.string(),
  cashbackPct: z.number(),
  estimatedAmount: z.number(),
  itemCount: z.number(),
})

export const CashbackSummarySchema = z.object({
  missionId: z.string(),
  merchantCashback: z.array(MerchantCashbackSchema).nullable().default([]),
  totalEstimated: z.number(),
  bestPlatform: z.string(),
  tip: z.string(),
})

export type MerchantCashback = z.infer<typeof MerchantCashbackSchema>
export type CashbackSummary = z.infer<typeof CashbackSummarySchema>

export async function getCashbackSummary(missionId: string): Promise<CashbackSummary> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/cashback`)
  return CashbackSummarySchema.parse(data)
}
