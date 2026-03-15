import { z } from 'zod'
import { apiFetch } from './client'

const MonthRecommendationSchema = z.object({
  month: z.number(),
  monthName: z.string(),
  score: z.number(),
  events: z.array(z.string()),
  advice: z.string(),
})

const CategoryCalendarSchema = z.object({
  category: z.string(),
  bestMonths: z.array(z.number()),
  worstMonths: z.array(z.number()),
  monthRecommendations: z.array(MonthRecommendationSchema),
  currentMonthScore: z.number(),
  currentAdvice: z.string(),
})

export type MonthRecommendation = z.infer<typeof MonthRecommendationSchema>
export type CategoryCalendar = z.infer<typeof CategoryCalendarSchema>

export async function getSeasonalCalendar(missionId: string): Promise<CategoryCalendar> {
  return apiFetch<CategoryCalendar>(`/api/missions/${missionId}/seasonal-calendar`)
}
