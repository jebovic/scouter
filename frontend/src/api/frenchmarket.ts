import { z } from 'zod'
import { apiFetch } from './client'

export const RetailerPricingSchema = z.object({
  name: z.string(),
  url: z.string(),
  typicalDiscount: z.string(),
  loyaltyProgram: z.string(),
})

export const PriceRangeSchema = z.object({
  min: z.number(),
  max: z.number(),
})

export const MarketInsightSchema = z.object({
  itemName: z.string(),
  category: z.string(),
  averagePriceEUR: z.number(),
  priceRange: PriceRangeSchema,
  bestRetailers: z.array(RetailerPricingSchema),
  bestMonthToBuy: z.string(),
  flashSaleFrequency: z.string(),
  tipsForFrenchBuyers: z.array(z.string()),
  marketTrend: z.string(),
})

export type MarketInsight = z.infer<typeof MarketInsightSchema>
export type RetailerPricing = z.infer<typeof RetailerPricingSchema>

export async function getMarketInsight(item: string, category: string): Promise<MarketInsight> {
  const params = new URLSearchParams({ item, category })
  const raw = await apiFetch<unknown>(`/api/market/insight?${params}`)
  return MarketInsightSchema.parse(raw)
}
