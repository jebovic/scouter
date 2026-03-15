import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const HolidayEventSchema = z.object({
  name: z.string(),
  date: z.string(),
  endDate: z.string().optional(),
  type: z.enum(['holiday', 'shopping_event']),
  daysUntil: z.number(),
  isActive: z.boolean(),
  advice: z.string(),
})
export type HolidayEvent = z.infer<typeof HolidayEventSchema>

const HolidayAlertResponseSchema = z.object({
  events: z.array(HolidayEventSchema),
  nextShoppingEvent: HolidayEventSchema.nullable().optional(),
  activeEvent: HolidayEventSchema.nullable().optional(),
  generatedAt: z.string(),
})
export type HolidayAlertResponse = z.infer<typeof HolidayAlertResponseSchema>

// ── API function ─────────────────────────────────────────────────────────────

export async function getHolidaysAndEvents(): Promise<HolidayAlertResponse> {
  const raw = await apiFetch<unknown>('/api/holidays-and-events')
  return HolidayAlertResponseSchema.parse(raw)
}
