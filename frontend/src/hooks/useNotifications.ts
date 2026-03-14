import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  listNotifications,
  getUnreadCount,
  markNotificationRead,
  markAllNotificationsRead,
} from '../api'
import type { Notification } from '../types'

const POLL_INTERVAL_MS = 60_000

export function useNotifications() {
  const { data: notifications = [], isLoading } = useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: () => listNotifications(),
    refetchInterval: POLL_INTERVAL_MS,
  })
  return { notifications, isLoading }
}

export function useUnreadCount() {
  const { data: count = 0 } = useQuery<number>({
    queryKey: ['notifications', 'unread-count'],
    queryFn: getUnreadCount,
    refetchInterval: POLL_INTERVAL_MS,
  })
  return count
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
