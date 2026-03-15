import { z } from 'zod'
import { apiFetch } from './client'

const VelocityReportSchema = z.object({
  missionId: z.string(),
  totalBudget: z.number(),
  totalSpent: z.number(),
  remainingBudget: z.number(),
  daysActive: z.number(),
  dailyVelocity: z.number(),
  projectedDaysLeft: z.number(),
  paceStatus: z.enum(['fast', 'moderate', 'on_track', 'slow', 'unconfigured']),
  paceLabel: z.string(),
  advice: z.string(),
  budgetUsedPercent: z.number(),
})

export type VelocityReport = z.infer<typeof VelocityReportSchema>

export async function getSpendingVelocity(missionId: string): Promise<VelocityReport> {
  return apiFetch<VelocityReport>(`/api/missions/${missionId}/spending-velocity`)
}
