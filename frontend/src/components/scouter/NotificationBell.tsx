import { useState, useRef, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useNotifications, useMarkNotificationRead, useMarkAllRead } from '../../hooks'
import styles from './NotificationBell.module.css'

export function NotificationBell() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const { notifications } = useNotifications()
  const unread = notifications.filter((n) => !n.read).length
  const markRead = useMarkNotificationRead()
  const { mutate: markAll } = useMarkAllRead()
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function mouseHandler(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    function keyHandler(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false)
    }
    if (open) {
      document.addEventListener('mousedown', mouseHandler)
      document.addEventListener('keydown', keyHandler)
    }
    return () => {
      document.removeEventListener('mousedown', mouseHandler)
      document.removeEventListener('keydown', keyHandler)
    }
  }, [open])

  return (
    <div className={styles.wrap} ref={ref}>
      <button
        className={styles.bell}
        onClick={() => setOpen((v) => !v)}
        aria-label={unread > 0 ? t('notifications.bellAriaLabelUnread', { count: unread }) : t('notifications.bellAriaLabel')}
      >
        🔔
        {unread > 0 && (
          <span className={styles.badge}>{unread > 9 ? '9+' : unread}</span>
        )}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Notifications">
          <div className={styles.header}>
            <span className={styles.title}>{t('notifications.alerts')}</span>
            {unread > 0 && (
              <button className={styles.markAll} onClick={() => markAll()}>
                {t('notifications.markAllRead')}
              </button>
            )}
          </div>

          {notifications.length === 0 ? (
            <p className={styles.empty}>{t('notifications.empty')}</p>
          ) : (
            <ul className={styles.list}>
              {notifications.slice(0, 5).map((n) => (
                <li
                  key={n.id}
                  className={`${styles.item} ${n.read ? styles.read : styles.unread}`}
                  role={!n.read ? 'button' : undefined}
                  tabIndex={!n.read ? 0 : undefined}
                  onClick={() => !n.read && markRead(n.id)}
                  onKeyDown={(e) => (e.key === 'Enter' || e.key === ' ') && !n.read && markRead(n.id)}
                >
                  <span className={styles.type}>{n.type.replaceAll('_', ' ')}</span>
                  <span className={styles.itemTitle}>{n.title}</span>
                  <span className={styles.body}>{n.body}</span>
                </li>
              ))}
            </ul>
          )}
          <div className={styles.footer}>
            <Link
              to="/notifications"
              className={styles.viewAll}
              onClick={() => setOpen(false)}
            >
              {t('notifications.viewAll')}
            </Link>
          </div>
        </div>
      )}
    </div>
  )
}
