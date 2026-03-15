import { z } from 'zod'
import { apiFetch } from './client'

export const ActivitySchema = z.object({
  type: z.string(),
  missionId: z.string(),
  missionName: z.string(),
  missionSlug: z.string(),
  itemId: z.string().optional(),
  itemName: z.string().optional(),
  description: z.string(),
  delta: z.number().optional(),
  occurredAt: z.string(),
})

export const ActivityFeedSchema = z.object({
  activities: z.array(ActivitySchema),
  totalMissions: z.number(),
  totalItems: z.number(),
  activeAlerts: z.number(),
  totalBudget: z.number(),
  generatedAt: z.string(),
})

export type Activity = z.infer<typeof ActivitySchema>
export type ActivityFeed = z.infer<typeof ActivityFeedSchema>

export async function getActivityFeed(): Promise<ActivityFeed> {
  const data = await apiFetch<unknown>('/api/activity-feed')
  return ActivityFeedSchema.parse(data)
}
