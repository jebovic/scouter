import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const WishlistItemSchema = z.object({
  name: z.string(),
  price: z.number(),
  status: z.string(),
  merchant: z.string(),
})

export const WishlistShareCardSchema = z.object({
  missionId: z.string(),
  missionName: z.string(),
  totalBudget: z.number(),
  currency: z.string(),
  items: z.array(WishlistItemSchema),
  shareUrl: z.string(),
  generatedAt: z.string(),
})

export type WishlistItem = z.infer<typeof WishlistItemSchema>
export type WishlistShareCard = z.infer<typeof WishlistShareCardSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function getWishlistCard(missionId: string): Promise<WishlistShareCard> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/wishlist-card`)
  return WishlistShareCardSchema.parse(data)
}
