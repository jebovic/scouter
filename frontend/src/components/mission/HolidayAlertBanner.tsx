import { useState } from 'react'
import { useHolidayAlerts } from '../../hooks/useHolidayAlerts'
import type { HolidayEvent } from '../../api/holidayalert'
import styles from './HolidayAlertBanner.module.css'

const DISMISS_PREFIX = 'holidayalert-dismissed-'

function isDismissed(eventName: string): boolean {
  try {
    return localStorage.getItem(`${DISMISS_PREFIX}${eventName}`) === '1'
  } catch {
    return false
  }
}

function setDismissed(eventName: string): void {
  try {
    localStorage.setItem(`${DISMISS_PREFIX}${eventName}`, '1')
  } catch {
    // ignore storage errors in restricted contexts
  }
}

function metaLabel(event: HolidayEvent): string {
  if (event.isActive) {
    if (event.endDate) {
      const end = new Intl.DateTimeFormat('fr-FR', {
        day: 'numeric',
        month: 'long',
      }).format(new Date(event.endDate))
      return `Actif jusqu'au ${end}`
    }
    return "Aujourd'hui"
  }

  const days = event.daysUntil
  if (days === 1) return 'Demain'
  if (days > 1) return `Dans ${days} jours`

  // Past events should not normally be shown, but handle gracefully.
  return `Il y a ${Math.abs(days)} jours`
}

interface BannerProps {
  event: HolidayEvent
  onDismiss: () => void
}

function Banner({ event, onDismiss }: BannerProps) {
  const icon = event.type === 'shopping_event' ? '🛍' : '📅'

  return (
    <div
      className={styles.banner}
      data-type={event.type}
      role="alert"
      aria-label={event.name}
    >
      <span className={styles.icon} aria-hidden="true">{icon}</span>
      <div className={styles.body}>
        <p className={styles.title}>{event.name}</p>
        <p className={styles.advice}>{event.advice}</p>
        <span className={styles.meta}>{metaLabel(event)}</span>
      </div>
      <button
        className={styles.dismiss}
        onClick={onDismiss}
        aria-label={`Masquer l'alerte pour ${event.name}`}
      >
        ✕
      </button>
    </div>
  )
}

export function HolidayAlertBanner() {
  const { activeEvent, nextShoppingEvent, isLoading } = useHolidayAlerts()
  const [dismissedKeys, setDismissedKeys] = useState<Set<string>>(new Set())

  if (isLoading) return null

  // Prefer the active event; fall back to the next upcoming shopping event.
  const candidate = activeEvent ?? nextShoppingEvent
  if (!candidate) return null

  const key = candidate.name
  if (isDismissed(key) || dismissedKeys.has(key)) return null

  function handleDismiss() {
    setDismissed(key)
    setDismissedKeys((prev) => new Set([...prev, key]))
  }

  return <Banner event={candidate} onDismiss={handleDismiss} />
}
