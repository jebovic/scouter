import { z } from 'zod'
import { apiFetch } from './client'

export const DigestAlertSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  missionName: z.string(),
  missionSlug: z.string(),
  oldPrice: z.number(),
  newPrice: z.number(),
  targetPrice: z.number().optional(),
  changePct: z.number(),
  alertType: z.enum(['big_drop', 'new_low', 'target_hit']),
})

export const WeeklyDigestSchema = z.object({
  weekStart: z.string(),
  weekEnd: z.string(),
  totalAlerts: z.number(),
  bigDrops: z.array(DigestAlertSchema),
  newLows: z.array(DigestAlertSchema),
  targetHits: z.array(DigestAlertSchema),
  summary: z.string(),
  generatedAt: z.string(),
})

export type DigestAlert = z.infer<typeof DigestAlertSchema>
export type WeeklyDigest = z.infer<typeof WeeklyDigestSchema>

export async function getWeeklyDigest(): Promise<WeeklyDigest> {
  const data = await apiFetch<unknown>('/api/weekly-digest')
  return WeeklyDigestSchema.parse(data)
}
