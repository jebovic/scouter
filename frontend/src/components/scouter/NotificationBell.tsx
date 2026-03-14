import { useState, useRef, useEffect } from 'react'
import { useNotifications, useUnreadCount, useMarkNotificationRead, useMarkAllRead } from '../../hooks'
import styles from './NotificationBell.module.css'

export function NotificationBell() {
  const [open, setOpen] = useState(false)
  const { notifications } = useNotifications()
  const unread = useUnreadCount()
  const markRead = useMarkNotificationRead()
  const markAll = useMarkAllRead()
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handler(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    if (open) document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [open])

  return (
    <div className={styles.wrap} ref={ref}>
      <button
        className={styles.bell}
        onClick={() => setOpen((v) => !v)}
        aria-label={`Notifications${unread > 0 ? ` (${unread} unread)` : ''}`}
      >
        🔔
        {unread > 0 && (
          <span className={styles.badge}>{unread > 9 ? '9+' : unread}</span>
        )}
      </button>

      {open && (
        <div className={styles.panel}>
          <div className={styles.header}>
            <span className={styles.title}>ALERTS</span>
            {unread > 0 && (
              <button className={styles.markAll} onClick={() => markAll()}>
                mark all read
              </button>
            )}
          </div>

          {notifications.length === 0 ? (
            <p className={styles.empty}>No notifications</p>
          ) : (
            <ul className={styles.list}>
              {notifications.map((n) => (
                <li
                  key={n.id}
                  className={`${styles.item} ${n.read ? styles.read : styles.unread}`}
                  onClick={() => !n.read && markRead(n.id)}
                >
                  <span className={styles.type}>{n.type.replace('_', ' ')}</span>
                  <span className={styles.itemTitle}>{n.title}</span>
                  <span className={styles.body}>{n.body}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}
