import { z } from 'zod'
import { apiFetch } from './client'

export const HeatmapCellSchema = z.object({
  category: z.string(),
  month: z.string(),
  total: z.number(),
  count: z.number().int(),
  intensity: z.number(),
})

export const BudgetHeatmapSchema = z.object({
  cells: z.array(HeatmapCellSchema),
  categories: z.array(z.string()),
  months: z.array(z.string()),
  maxTotal: z.number(),
  generatedAt: z.string(),
})

export type HeatmapCell = z.infer<typeof HeatmapCellSchema>
export type BudgetHeatmap = z.infer<typeof BudgetHeatmapSchema>

export async function getBudgetHeatmap(): Promise<BudgetHeatmap> {
  const data = await apiFetch<unknown>('/api/analytics/budget-heatmap')
  return BudgetHeatmapSchema.parse(data)
}
