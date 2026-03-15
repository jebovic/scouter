import { z } from 'zod'
import { apiFetch } from './client'

export const CategoryStatsSchema = z.object({
  category: z.string(),
  total: z.number(),
  itemCount: z.number().int(),
  avgPrice: z.number(),
  pct: z.number(),
})

export const MerchantStatsSchema = z.object({
  merchant: z.string(),
  total: z.number(),
  itemCount: z.number().int(),
})

export const TopItemSchema = z.object({
  itemId: z.string(),
  itemName: z.string(),
  missionName: z.string(),
  price: z.number(),
  category: z.string(),
})

export const MonthStatsSchema = z.object({
  month: z.string(),
  avgPrice: z.number(),
  dataPoints: z.number().int(),
})

export const SpendingAnalyticsSchema = z.object({
  totalBudget: z.number(),
  totalSpending: z.number(),
  budgetUsagePct: z.number(),
  byCategory: z.array(CategoryStatsSchema),
  byMerchant: z.array(MerchantStatsSchema),
  topItems: z.array(TopItemSchema),
  monthlySummary: z.array(MonthStatsSchema),
  generatedAt: z.string(),
})

export type SpendingAnalytics = z.infer<typeof SpendingAnalyticsSchema>
export type CategoryStats = z.infer<typeof CategoryStatsSchema>
export type MerchantStats = z.infer<typeof MerchantStatsSchema>
export type TopItem = z.infer<typeof TopItemSchema>
export type MonthStats = z.infer<typeof MonthStatsSchema>

export async function getSpendingAnalytics(): Promise<SpendingAnalytics> {
  const data = await apiFetch<unknown>('/api/analytics/spending')
  return SpendingAnalyticsSchema.parse(data)
}
