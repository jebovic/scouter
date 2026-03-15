import { z } from 'zod'
import { apiFetch } from './client'

export const PriceDropItemSchema = z.object({
  missionId: z.string(),
  missionName: z.string(),
  missionSlug: z.string(),
  itemId: z.string(),
  itemName: z.string(),
  merchant: z.string(),
  priceBefore: z.number(),
  priceNow: z.number(),
  dropPct: z.number(),
  currency: z.string(),
})

export const WeeklyDigestSchema = z.object({
  generatedAt: z.string(),
  periodDays: z.number(),
  totalDrops: z.number(),
  totalSavings: z.number(),
  currency: z.string(),
  items: z.array(PriceDropItemSchema),
})

export type WeeklyDigest = z.infer<typeof WeeklyDigestSchema>
export type PriceDropItem = z.infer<typeof PriceDropItemSchema>

export async function fetchWeeklyDigest(days = 7): Promise<WeeklyDigest> {
  const data = await apiFetch<unknown>(`/api/digest/weekly?days=${days}`)
  return WeeklyDigestSchema.parse(data)
}
