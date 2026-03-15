import { z } from 'zod'
import { apiFetch } from './client'

export const CompetitorOfferSchema = z.object({
  platform: z.string(),
  price: z.number(),
  priceDiff: z.number(),
  priceDiffPct: z.number(),
  url: z.string(),
  isBest: z.boolean(),
  availability: z.string(),
})

export const CompetitorPricesSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  basePrice: z.number(),
  competitors: z.array(CompetitorOfferSchema),
  bestOffer: CompetitorOfferSchema.optional(),
  savings: z.number(),
  generatedAt: z.string(),
})

export type CompetitorOffer = z.infer<typeof CompetitorOfferSchema>
export type CompetitorPrices = z.infer<typeof CompetitorPricesSchema>

export async function getCompetitorPrices(
  missionId: string,
  itemId: string,
): Promise<CompetitorPrices> {
  const raw = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/competitor-prices`,
  )
  return CompetitorPricesSchema.parse(raw)
}
