import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { fetchComments, createComment, deleteComment } from '../api/comment'

export function useComments(missionSlug: string, optionId?: string) {
  return useQuery({
    queryKey: ['comments', missionSlug, optionId],
    queryFn: () => fetchComments(missionSlug, optionId),
    enabled: !!missionSlug,
    staleTime: 30 * 1000,
  })
}

export function useCreateComment(missionSlug: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: { optionId?: string; author: string; body: string }) =>
      createComment(missionSlug, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', missionSlug] })
    },
  })
}

export function useDeleteComment(missionSlug: string) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (commentId: string) => deleteComment(missionSlug, commentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['comments', missionSlug] })
    },
  })
}
