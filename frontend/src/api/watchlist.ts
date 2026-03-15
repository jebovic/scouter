import { z } from 'zod'
import { apiFetch } from './client'

export const WatchEntrySchema = z.object({
  id: z.string(),
  itemId: z.string(),
  itemName: z.string(),
  baselinePrice: z.number(),
  thresholdPct: z.number(),
  currentPrice: z.number(),
  triggeredAt: z.string().nullable().optional(),
  createdAt: z.string(),
})

export const CheckResultSchema = z.object({
  triggered: z.boolean(),
  dropPct: z.number(),
  message: z.string(),
})

export type WatchEntry = z.infer<typeof WatchEntrySchema>

export async function getWatchlist(): Promise<WatchEntry[]> {
  const raw = await apiFetch<unknown[]>('/api/watchlist')
  return raw.map((r) => WatchEntrySchema.parse(r))
}

export async function addToWatchlist(req: {
  itemId: string
  itemName: string
  baselinePrice: number
  thresholdPct: number
}): Promise<WatchEntry> {
  const raw = await apiFetch<unknown>('/api/watchlist', {
    method: 'POST',
    body: JSON.stringify(req),
  })
  return WatchEntrySchema.parse(raw)
}

export async function removeFromWatchlist(id: string): Promise<void> {
  await apiFetch<void>(`/api/watchlist/${id}`, { method: 'DELETE' })
}
