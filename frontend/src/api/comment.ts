import { z } from 'zod'
import { apiFetch } from './client'

export const CommentSchema = z.object({
  id: z.string(),
  missionId: z.string(),
  optionId: z.string().optional(),
  author: z.string(),
  body: z.string(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Comment = z.infer<typeof CommentSchema>

export async function fetchComments(missionSlug: string, optionId?: string): Promise<Comment[]> {
  const params = optionId ? `?optionId=${encodeURIComponent(optionId)}` : ''
  const data = await apiFetch<unknown>(`/api/missions/${missionSlug}/comments${params}`)
  return z.array(CommentSchema).parse(data)
}

export async function createComment(
  missionSlug: string,
  payload: { optionId?: string; author: string; body: string },
): Promise<Comment> {
  const data = await apiFetch<unknown>(`/api/missions/${missionSlug}/comments`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
  return CommentSchema.parse(data)
}

export async function deleteComment(missionSlug: string, commentId: string): Promise<void> {
  await apiFetch<void>(`/api/missions/${missionSlug}/comments/${commentId}`, { method: 'DELETE' })
}
