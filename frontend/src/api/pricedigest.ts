import { z } from 'zod'
import { apiFetch } from './client'

export const PriceChangeSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  merchant: z.string(),
  missionName: z.string(),
  missionSlug: z.string(),
  oldPrice: z.number(),
  newPrice: z.number(),
  changeAmount: z.number(),
  changePct: z.number(),
  status: z.string(),
  isDropping: z.boolean(),
})

export const DigestSchema = z.object({
  generatedAt: z.string(),
  totalItemsChecked: z.number(),
  priceDrops: z.array(PriceChangeSchema),
  priceRises: z.array(PriceChangeSchema),
  stableItems: z.number(),
  biggestDrop: PriceChangeSchema.nullable(),
  biggestRise: PriceChangeSchema.nullable(),
  summary: z.string(),
})

export type PriceChange = z.infer<typeof PriceChangeSchema>
export type Digest = z.infer<typeof DigestSchema>

export async function getPriceDigest(): Promise<Digest> {
  const data = await apiFetch<unknown>('/api/price-digest')
  return DigestSchema.parse(data)
}
