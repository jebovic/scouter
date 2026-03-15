import { z } from 'zod'

const BundleItemSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  price: z.number(),
})

const BundleDealSchema = z.object({
  items: z.array(BundleItemSchema),
  bundleName: z.string(),
  totalPrice: z.number(),
  estimatedSave: z.number(),
  savePercent: z.number(),
  tip: z.string(),
  category: z.string(),
  confidence: z.enum(['high', 'medium', 'low']),
})

export const BundleResponseSchema = z.object({
  missionId: z.string(),
  bundles: z.array(BundleDealSchema),
  totalPotentialSave: z.number(),
  message: z.string(),
})

export type BundleDeal = z.infer<typeof BundleDealSchema>
export type BundleResponse = z.infer<typeof BundleResponseSchema>

export async function getBundleDeals(missionId: string): Promise<BundleResponse> {
  const res = await fetch(`/api/missions/${missionId}/bundle-deals`)
  if (!res.ok) throw new Error('Failed to fetch bundle deals')
  return BundleResponseSchema.parse(await res.json())
}
