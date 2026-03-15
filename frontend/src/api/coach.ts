import { z } from 'zod'
import { apiFetch } from './client'

export const CoachTipSchema = z.object({
  id: z.string(),
  category: z.enum(['research', 'budget', 'timing', 'decision']),
  priority: z.enum(['high', 'medium', 'low']),
  title: z.string(),
  body: z.string(),
  action: z.string().optional(),
})

export const CoachResponseSchema = z.object({
  missionId: z.string(),
  tips: z.array(CoachTipSchema),
  generatedAt: z.string(),
})

export type CoachTip = z.infer<typeof CoachTipSchema>
export type CoachResponse = z.infer<typeof CoachResponseSchema>

export async function fetchCoachTips(missionSlug: string): Promise<CoachResponse> {
  const data = await apiFetch<unknown>(`/api/missions/${missionSlug}/coach`)
  return CoachResponseSchema.parse(data)
}
