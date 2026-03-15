import { z } from 'zod'
import { apiFetch } from './client'

const DealSchema = z.object({
  source: z.string(),
  title: z.string(),
  origPrice: z.number(),
  dealPrice: z.number(),
  discount: z.number(),
  url: z.string(),
  hot: z.boolean(),
  votes: z.number(),
})

const DealAggregatorResponseSchema = z.object({
  itemId: z.string(),
  deals: z.array(DealSchema),
  bestDeal: DealSchema.nullable().optional(),
})

export type Deal = z.infer<typeof DealSchema>
export type DealAggregatorResponse = z.infer<typeof DealAggregatorResponseSchema>

export async function getDeals(missionId: string, itemId: string): Promise<DealAggregatorResponse> {
  const raw = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/deals`)
  return DealAggregatorResponseSchema.parse(raw)
}
