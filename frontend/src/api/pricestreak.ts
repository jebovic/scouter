import { z } from 'zod'
import { apiFetch } from './client'

export const PriceStreakSchema = z.object({
  itemId: z.string(),
  currentStreak: z.number().int(),
  streakType: z.enum(['dropping', 'rising', 'stable']),
  totalDropPct: z.number(),
  totalDropAmt: z.number(),
  streakStart: z.string(),
  streakStartPrice: z.number(),
  recommendation: z.string(),
  urgency: z.enum(['low', 'medium', 'high', 'urgent']),
  generatedAt: z.string(),
})

export type PriceStreak = z.infer<typeof PriceStreakSchema>

export async function getPriceStreak(missionId: string, itemId: string): Promise<PriceStreak> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/items/${itemId}/price-streak`)
  return PriceStreakSchema.parse(data)
}
