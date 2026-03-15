import { z } from 'zod'
import { apiFetch } from './client'

export const CarbonEstimateSchema = z.object({
  itemName: z.string(),
  category: z.string(),
  kgCO2: z.number(),
  label: z.string(),
  tips: z.array(z.string()),
  confidence: z.enum(['high', 'medium', 'low']),
})

export type CarbonEstimate = z.infer<typeof CarbonEstimateSchema>

export async function getCarbonEstimate(
  itemName: string,
  category: string,
): Promise<CarbonEstimate> {
  const params = new URLSearchParams({ itemName, category })
  const raw = await apiFetch<unknown>(`/api/carbon?${params}`)
  return CarbonEstimateSchema.parse(raw)
}
