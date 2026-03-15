import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listNotifications,
  markNotificationRead,
  markAllNotificationsRead,
  deleteNotification,
} from '../api'
import type { Notification, NotificationFilterParams } from '../api/notifications'

const POLL_INTERVAL_MS = 60_000

export function useNotifications(params: NotificationFilterParams = {}) {
  const { data: notifications = [], isLoading, error } = useQuery<Notification[]>({
    queryKey: ['notifications', params],
    queryFn: () => listNotifications(params),
    refetchInterval: POLL_INTERVAL_MS,
  })
  return { notifications, isLoading, error }
}

// Derived from the notifications list — no separate network request.
export function useUnreadCount() {
  const { notifications } = useNotifications()
  return notifications.filter((n) => !n.read).length
}

export function useMarkNotificationRead() {
  const qc = useQueryClient()
  const { mutate } = useMutation({
    mutationFn: markNotificationRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
  return mutate
}

export function useMarkAllRead() {
  const qc = useQueryClient()
  const { mutate, isPending } = useMutation({
    mutationFn: markAllNotificationsRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
  return { mutate, isPending }
}

export function useDeleteNotification() {
  const qc = useQueryClient()
  const { mutate, isPending } = useMutation({
    mutationFn: deleteNotification,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
  return { mutate, isPending }
}
