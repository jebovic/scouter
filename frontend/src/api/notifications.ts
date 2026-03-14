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

// ── API functions ────────────────────────────────────────────────────────────

export async function listNotifications(limit = 50): Promise<Notification[]> {
  const data = await apiFetch<unknown>(`/api/notifications?limit=${limit}`)
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
