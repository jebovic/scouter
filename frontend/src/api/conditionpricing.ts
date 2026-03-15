import { z } from 'zod'
import { apiFetch } from './client'

export const ConditionPriceSchema = z.object({
  condition: z.string(),
  grade: z.string(),
  estPrice: z.number(),
  discount: z.number(),
  platform: z.string(),
  availability: z.string(),
  warrantyMonths: z.number(),
  searchUrl: z.string(),
})

export const ConditionReportSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  newPrice: z.number(),
  conditions: z.array(ConditionPriceSchema),
  bestDeal: ConditionPriceSchema.optional(),
  recommendation: z.string(),
  generatedAt: z.string(),
})

export type ConditionPrice = z.infer<typeof ConditionPriceSchema>
export type ConditionReport = z.infer<typeof ConditionReportSchema>

export async function getConditionPricing(
  missionId: string,
  itemId: string,
): Promise<ConditionReport> {
  const raw = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/condition-pricing`,
  )
  return ConditionReportSchema.parse(raw)
}
