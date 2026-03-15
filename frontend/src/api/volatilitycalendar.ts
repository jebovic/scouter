import { z } from 'zod'
import { apiFetch } from './client'

export const WeekDataSchema = z.object({
  week: z.number().int(),
  year: z.number().int(),
  intensity: z.number(),
  minPrice: z.number(),
  maxPrice: z.number(),
  label: z.string(),
})

export const VolatilityCalendarSchema = z.object({
  itemId: z.string(),
  weeks: z.array(WeekDataSchema),
  mostVolatileWeek: WeekDataSchema.nullable(),
  recommendation: z.string(),
})

export type WeekData = z.infer<typeof WeekDataSchema>
export type VolatilityCalendar = z.infer<typeof VolatilityCalendarSchema>

export async function getVolatilityCalendar(
  missionId: string,
  itemId: string,
): Promise<VolatilityCalendar> {
  const data = await apiFetch<unknown>(
    `/api/missions/${missionId}/items/${itemId}/volatility-calendar`,
  )
  return VolatilityCalendarSchema.parse(data)
}
