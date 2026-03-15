import { useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useMissions } from '../hooks/useMission'
import {
  useNotifications,
  useMarkNotificationRead,
  useMarkAllRead,
  useDeleteNotification,
} from '../hooks'
import type { Notification } from '../api/notifications'
import styles from './NotificationsPage.module.css'

type FilterType = 'all' | 'price_drop' | 'target_hit' | 'trend_alert'

const TYPE_ICON: Record<string, string> = {
  price_drop: '📉',
  target_hit: '🎯',
  trend_alert: '📈',
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60_000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  const days = Math.floor(hrs / 24)
  return `${days}d ago`
}

export default function NotificationsPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [filter, setFilter] = useState<FilterType>('all')

  const filterParam = filter === 'all' ? undefined : filter
  const { notifications, isLoading } = useNotifications(
    filterParam ? { type: filterParam } : {}
  )
  const { missions } = useMissions()
  const markRead = useMarkNotificationRead()
  const { mutate: markAll } = useMarkAllRead()
  const { mutate: deleteNotif } = useDeleteNotification()

  const allRead = notifications.every((n) => n.read)

  const missionNameMap = Object.fromEntries(
    missions.map((m) => [m.id, m.name])
  )

  const missionSlugMap = Object.fromEntries(
    missions.map((m) => [m.id, m.slug])
  )

  // Group by missionId
  const grouped = notifications.reduce<Record<string, Notification[]>>((acc, n) => {
    const key = n.missionId
    if (!acc[key]) acc[key] = []
    acc[key].push(n)
    return acc
  }, {})

  const missionIds = Object.keys(grouped)

  const handleRowClick = useCallback(
    (n: Notification) => {
      if (!n.read) markRead(n.id)
      const slug = missionSlugMap[n.missionId] ?? n.missionId
      navigate(`/missions/${slug}`)
    },
    [markRead, navigate, missionSlugMap]
  )

  const handleDelete = useCallback(
    (e: React.MouseEvent, id: string) => {
      e.stopPropagation()
      deleteNotif(id)
    },
    [deleteNotif]
  )

  const tabs: { key: FilterType; label: string }[] = [
    { key: 'all', label: t('notifications.filterAll') },
    { key: 'price_drop', label: t('notifications.filterPriceDrop') },
    { key: 'target_hit', label: t('notifications.filterTargetHit') },
    { key: 'trend_alert', label: t('notifications.filterTrendAlert') },
  ]

  return (
    <main className={styles.page}>
      <div className={styles.header}>
        <h1 className={styles.title}>{t('notifications.title').toUpperCase()}</h1>
        <button
          className={styles.markAllBtn}
          disabled={allRead || isLoading}
          onClick={() => markAll()}
        >
          {t('notifications.markAllRead')}
        </button>
      </div>

      <div className={styles.filterBar} role="tablist" aria-label="Filter notifications">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            role="tab"
            aria-selected={filter === tab.key}
            className={`${styles.filterTab} ${filter === tab.key ? styles.filterTabActive : ''}`}
            onClick={() => setFilter(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {!isLoading && notifications.length === 0 && (
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>🔔</div>
          <div className={styles.emptyTitle}>{t('notifications.empty')}</div>
          <div className={styles.emptyDesc}>{t('notifications.emptyDesc')}</div>
        </div>
      )}

      {missionIds.map((missionId) => {
        const items = grouped[missionId]
        const name = missionNameMap[missionId] ?? missionId
        return (
          <section key={missionId} className={styles.missionGroup}>
            <div className={styles.missionHeader}>{name}</div>
            <ul className={styles.notifList}>
              {items.map((n) => (
                <li
                  key={n.id}
                  className={`${styles.notifRow} ${n.read ? styles.notifRowRead : styles.notifRowUnread}`}
                >
                  <button
                    className={styles.notifRowBtn}
                    onClick={() => handleRowClick(n)}
                    aria-label={n.title}
                  >
                    <span className={styles.typeIcon} aria-hidden="true">
                      {TYPE_ICON[n.type] ?? '🔔'}
                    </span>
                    <div className={styles.notifContent}>
                      <div className={styles.notifTitle}>{n.title}</div>
                      <div className={styles.notifBody}>{n.body}</div>
                      <div className={styles.notifMeta}>
                        <span className={styles.notifTime}>{relativeTime(n.createdAt)}</span>
                        {!n.read && <span className={styles.unreadDot} aria-label="unread" />}
                      </div>
                    </div>
                  </button>
                  <button
                    className={styles.deleteBtn}
                    title={t('notifications.deleteConfirm')}
                    aria-label={t('notifications.deleteConfirm')}
                    onClick={(e) => handleDelete(e, n.id)}
                  >
                    🗑️
                  </button>
                </li>
              ))}
            </ul>
          </section>
        )
      })}
    </main>
  )
}
