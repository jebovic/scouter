import { z } from 'zod'
import { apiFetch } from './client'

export const PriceFloorSchema = z.object({
  itemId: z.string(),
  currentPrice: z.number(),
  predictedFloor: z.number(),
  floorConfidence: z.enum(['low', 'medium', 'high']),
  historicalMin: z.number(),
  distanceToFloor: z.number(),
  distancePct: z.number(),
  shouldWait: z.boolean(),
  waitReason: z.string(),
  buyNowReason: z.string(),
  generatedAt: z.string(),
})

export type PriceFloor = z.infer<typeof PriceFloorSchema>

export async function getPriceFloor(missionId: string, itemId: string): Promise<PriceFloor> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/price-floor`)
  return PriceFloorSchema.parse(data)
}
