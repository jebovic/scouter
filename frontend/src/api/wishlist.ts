import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ──────────────────────────────────────────────────────────────

export const WishListItemSchema = z.object({
  id: z.string(),
  name: z.string(),
  url: z.string().nullable(),
  targetPrice: z.number().nullable(),
  currency: z.string(),
  notes: z.string().nullable(),
  alertEnabled: z.boolean(),
  lastCheckedAt: z.string().nullable(),
  lastCheckedPrice: z.number().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type WishListItem = z.infer<typeof WishListItemSchema>

export const WishListSchema = z.array(WishListItemSchema)

// ── API functions ────────────────────────────────────────────────────────────

export async function fetchWishList(): Promise<WishListItem[]> {
  const data = await apiFetch<unknown>('/api/wishlist')
  return WishListSchema.parse(data)
}

export async function createWishListItem(payload: {
  name: string
  url?: string
  targetPrice?: number
  currency?: string
  notes?: string
}): Promise<WishListItem> {
  const data = await apiFetch<unknown>('/api/wishlist', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return WishListItemSchema.parse(data)
}

export async function deleteWishListItem(id: string): Promise<void> {
  await apiFetch<void>(`/api/wishlist/${id}`, { method: 'DELETE' })
}

export async function toggleWishListAlert(id: string, enabled: boolean): Promise<void> {
  await apiFetch<void>(`/api/wishlist/${id}/alert`, {
    method: 'PATCH',
    body: JSON.stringify({ alertEnabled: enabled }),
  })
}
