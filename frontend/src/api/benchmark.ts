import { z } from 'zod'
import { apiFetch } from './client'

export const PriceRangeSchema = z.object({
  low: z.number(),
  average: z.number(),
  high: z.number(),
})

export const BenchmarkResultSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  currentPrice: z.number(),
  currency: z.string(),
  marketRange: PriceRangeSchema,
  verdict: z.enum(['good_deal', 'average', 'overpriced']),
  diffPercent: z.number(),
  explanation: z.string(),
  cachedAt: z.number(),
})

export type PriceRange = z.infer<typeof PriceRangeSchema>
export type BenchmarkResult = z.infer<typeof BenchmarkResultSchema>

export async function getBenchmark(itemId: string): Promise<BenchmarkResult> {
  const data = await apiFetch<unknown>(`/api/shopping-items/${itemId}/benchmark`)
  return BenchmarkResultSchema.parse(data)
}
