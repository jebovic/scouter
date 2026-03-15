import { z } from 'zod'
import { apiFetch } from './client'

export const CategorySuggestionSchema = z.object({
  category: z.string(),
  confidence: z.number(),
  alternatives: z.array(z.string()).optional().default([]),
})

export type CategorySuggestion = z.infer<typeof CategorySuggestionSchema>

export async function suggestCategory(missionSlug: string): Promise<CategorySuggestion> {
  const data = await apiFetch<unknown>(`/api/missions/${missionSlug}/suggest-category`, {
    method: 'POST',
  })
  return CategorySuggestionSchema.parse(data)
}
