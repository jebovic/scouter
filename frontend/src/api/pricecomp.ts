import { z } from 'zod'
import { apiFetch } from './client'

export const RetailerOfferSchema = z.object({
  retailer: z.string(),
  price: z.number(),
  currency: z.string(),
  available: z.boolean(),
  url: z.string(),
  updatedAt: z.string(),
})

export const PriceComparisonSchema = z.object({
  optionId: z.string(),
  product: z.string(),
  offers: z.array(RetailerOfferSchema),
  lowestIdx: z.number(),
  fetchedAt: z.string(),
})

export type RetailerOffer = z.infer<typeof RetailerOfferSchema>
export type PriceComparison = z.infer<typeof PriceComparisonSchema>

export async function getPriceComparison(optionId: string): Promise<PriceComparison> {
  const data = await apiFetch<unknown>(`/api/options/${optionId}/price-comparison`)
  return PriceComparisonSchema.parse(data)
}

export async function refreshPriceComparison(optionId: string): Promise<PriceComparison> {
  const data = await apiFetch<unknown>(`/api/options/${optionId}/price-comparison/refresh`, {
    method: 'POST',
  })
  return PriceComparisonSchema.parse(data)
}
