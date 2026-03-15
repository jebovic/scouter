import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ───────────────────────────────────────────────────────────────

export const InflationItemSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  price: z.number(),
  annualInflationRate: z.number(),
  inflationPer30Days: z.number(),
  waitCost3Months: z.number(),
  recommendation: z.string(),
})

export const InflationImpactSchema = z.object({
  missionId: z.string(),
  items: z.array(InflationItemSchema),
  totalWaitCost3Months: z.number(),
  avgInflationRate: z.number(),
})

export type InflationItem = z.infer<typeof InflationItemSchema>
export type InflationImpact = z.infer<typeof InflationImpactSchema>

// ── API function ──────────────────────────────────────────────────────────────

export async function fetchInflationImpact(missionId: string): Promise<InflationImpact> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/inflation-impact`)
  return InflationImpactSchema.parse(data)
}
