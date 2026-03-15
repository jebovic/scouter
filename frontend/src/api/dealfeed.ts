import { z } from 'zod'
import { apiFetch } from './client'

export const PromoOfferSchema = z.object({
  title: z.string(),
  retailer: z.string(),
  originalPrice: z.number(),
  promoPrice: z.number(),
  discount: z.number(),
  url: z.string(),
  expiresAt: z.string().nullish(),
  category: z.string(),
  isFlash: z.boolean(),
  source: z.string(),
})

export const PromoFeedResultSchema = z.object({
  offers: z.array(PromoOfferSchema),
  query: z.string(),
  category: z.string(),
  fetchedAt: z.string(),
})

export type PromoOffer = z.infer<typeof PromoOfferSchema>
export type PromoFeedResult = z.infer<typeof PromoFeedResultSchema>

export async function getPromoFeed(q: string, category?: string): Promise<PromoFeedResult> {
  const params = new URLSearchParams({ q })
  if (category) params.set('category', category)
  const raw = await apiFetch<unknown>(`/api/promo-feed?${params}`)
  return PromoFeedResultSchema.parse(raw)
}
