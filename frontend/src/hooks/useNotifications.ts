import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listNotifications,
  markNotificationRead,
  markAllNotificationsRead,
} from '../api'
import type { Notification } from '../types'

const POLL_INTERVAL_MS = 60_000

export function useNotifications() {
  const { data: notifications = [], isLoading, error } = useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: () => listNotifications(),
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
  const { mutate } = useMutation({
    mutationFn: markAllNotificationsRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['notifications'] })
    },
  })
  return mutate
}
