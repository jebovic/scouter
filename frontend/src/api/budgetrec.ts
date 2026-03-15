import { z } from 'zod'
import { apiFetch } from './client'

const BudgetRecommendationSchema = z.object({
  category: z.string(),
  currentBudget: z.number(),
  recommendedBudget: z.number(),
  reason: z.string(),
  priority: z.enum(['high', 'medium', 'low']),
  saveAmount: z.number(),
})

export const BudgetAnalysisSchema = z.object({
  missionId: z.string(),
  totalBudget: z.number(),
  totalSpent: z.number(),
  recommendations: z.array(BudgetRecommendationSchema),
  potentialSavings: z.number(),
  healthScore: z.number(),
})

export type BudgetAnalysis = z.infer<typeof BudgetAnalysisSchema>
export type BudgetRecommendation = z.infer<typeof BudgetRecommendationSchema>

export async function getBudgetAnalysis(missionId: string): Promise<BudgetAnalysis> {
  const data = await apiFetch<unknown>(`/api/missions/${missionId}/budget-analysis`)
  return BudgetAnalysisSchema.parse(data)
}
