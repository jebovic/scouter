import { z } from 'zod'
import { apiFetch } from './client'

// ── Zod schemas ─────────────────────────────────────────────────────────────

export const NotificationSchema = z.object({
  id: z.string().uuid(),
  missionId: z.string().uuid(),
  itemId: z.string().uuid().nullish(),
  type: z.enum(['price_drop', 'target_hit', 'trend_alert']),
  title: z.string(),
  body: z.string(),
  read: z.boolean(),
  createdAt: z.string(),
})

export type Notification = z.infer<typeof NotificationSchema>

export const NotificationFilterParamsSchema = z.object({
  missionId: z.string().uuid().optional(),
  type: z.enum(['price_drop', 'target_hit', 'trend_alert']).optional(),
  limit: z.number().int().positive().optional(),
})

export type NotificationFilterParams = z.infer<typeof NotificationFilterParamsSchema>

// ── API functions ────────────────────────────────────────────────────────────

export async function listNotifications(params: NotificationFilterParams = {}): Promise<Notification[]> {
  const qs = new URLSearchParams()
  if (params.limit !== undefined) qs.set('limit', String(params.limit))
  if (params.missionId) qs.set('missionId', params.missionId)
  if (params.type) qs.set('type', params.type)
  const query = qs.toString()
  const data = await apiFetch<unknown>(`/api/notifications${query ? `?${query}` : ''}`)
  return z.array(NotificationSchema).parse(data)
}

export async function getUnreadCount(): Promise<number> {
  const data = await apiFetch<unknown>('/api/notifications/unread-count')
  const { count } = z.object({ count: z.number() }).parse(data)
  return count
}

export async function markNotificationRead(id: string): Promise<void> {
  await apiFetch<void>(`/api/notifications/${id}/read`, { method: 'PATCH' })
}

export async function markAllNotificationsRead(): Promise<void> {
  await apiFetch<void>('/api/notifications/read-all', { method: 'POST' })
}

export async function deleteNotification(id: string): Promise<void> {
  await apiFetch<void>(`/api/notifications/${id}`, { method: 'DELETE' })
}
