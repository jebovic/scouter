import { z } from 'zod'
import { apiFetch } from './client'

const RetailerPriceSchema = z.object({
  retailer: z.string(),
  price: z.number(),
  url: z.string(),
  isBestPrice: z.boolean(),
  diffPercent: z.number(),
})

const PriceComparisonSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  basePrice: z.number(),
  retailers: z.array(RetailerPriceSchema),
  bestDeal: RetailerPriceSchema.nullable(),
  maxSaving: z.number(),
})

export type PriceComparison = z.infer<typeof PriceComparisonSchema>

export async function getPriceComparison(missionId: string, itemId: string): Promise<PriceComparison> {
  const data = await apiFetch(`/api/missions/${missionId}/items/${itemId}/price-comparison`)
  return PriceComparisonSchema.parse(data)
}
