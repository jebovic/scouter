import { z } from 'zod'
import { apiFetch } from './client'

export const MarketplaceResultSchema = z.object({
  site: z.string(),
  url: z.string(),
  price: z.number().optional(),
  available: z.boolean(),
  label: z.string(),
})

export const MarketplaceResponseSchema = z.object({
  itemName: z.string(),
  results: z.array(MarketplaceResultSchema),
})

export type MarketplaceResult = z.infer<typeof MarketplaceResultSchema>
export type MarketplaceResponse = z.infer<typeof MarketplaceResponseSchema>

export async function getMarketplaceResults(query: string): Promise<MarketplaceResponse> {
  const params = new URLSearchParams({ q: query })
  const raw = await apiFetch<unknown>(`/api/marketplace?${params}`)
  return MarketplaceResponseSchema.parse(raw)
}
