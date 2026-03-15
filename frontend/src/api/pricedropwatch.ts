import { z } from 'zod'
import { apiFetch } from './client'

const WatchItemSchema = z.object({
  itemId: z.string(),
  name: z.string(),
  currentPrice: z.number(),
  dropProbability: z.number(), // 0-100 percentage
  signal: z.enum(['high_volatility', 'recent_spike', 'no_history', 'stable']),
  expectedDrop: z.number(), // percentage
  reasoning: z.string(),
})

const WatchlistResponseSchema = z.object({
  summary: z.string(),
  watchItems: z.array(WatchItemSchema),
})

export type WatchItem = z.infer<typeof WatchItemSchema>
export type WatchlistResponse = z.infer<typeof WatchlistResponseSchema>

export async function getPriceDropWatchlist(missionId: string): Promise<WatchlistResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/price-drop-watchlist`)
  return WatchlistResponseSchema.parse(data) as WatchlistResponse
}
