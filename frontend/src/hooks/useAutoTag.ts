import { useMutation } from '@tanstack/react-query'
import { suggestCategory } from '../api/autotag'
import type { CategorySuggestion } from '../api/autotag'

export function useSuggestCategory(missionSlug: string) {
  return useMutation<CategorySuggestion, Error>({
    mutationFn: () => suggestCategory(missionSlug),
  })
}
