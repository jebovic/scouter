import { z } from 'zod'
import { apiFetch } from './client'

const DataPointSchema = z.object({
  date: z.string(),
  amount: z.number(),
})

const BurnRateResponseSchema = z.object({
  missionId: z.string(),
  totalBudget: z.number(),
  totalSpent: z.number(),
  dailyRate: z.number(),
  projectedTotal: z.number(),
  daysElapsed: z.number(),
  burnPoints: z.array(DataPointSchema),
  status: z.enum(['on_track', 'at_risk', 'over_budget']),
})

export type BurnRateResponse = z.infer<typeof BurnRateResponseSchema>

export async function getBurnRate(missionId: string): Promise<BurnRateResponse> {
  return apiFetch(`/api/missions/${missionId}/burn-rate`, BurnRateResponseSchema)
}
