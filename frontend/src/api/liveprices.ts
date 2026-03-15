import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

const LivePriceListingSchema = z.object({
  source: z.string(),
  productName: z.string(),
  brand: z.string(),
  price: z.number(),
  currency: z.string(),
  url: z.string(),
  imageUrl: z.string(),
  condition: z.string(),
  stores: z.string(),
  updatedAt: z.string(),
})

const LivePricesResponseSchema = z.object({
  listings: z.array(LivePriceListingSchema),
  providers: z.array(z.string()),
  cached: z.boolean(),
})

export type LivePriceListing = z.infer<typeof LivePriceListingSchema>
export type LivePricesResponse = z.infer<typeof LivePricesResponseSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function fetchLivePrices(missionId: string, optionId: string): Promise<LivePricesResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/options/${optionId}/live-prices`)
  return LivePricesResponseSchema.parse(data)
}
