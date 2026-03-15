import { useSaleCalendar } from '../../hooks/useSaleCalendar'
import type { SaleEvent } from '../../api/salecalendar'
import styles from './SaleCalendarWidget.module.css'

const TYPE_ICON: Record<string, string> = {
  soldes: '🛍️',
  flash: '⚡',
  fete: '💝',
  back_to_school: '🎒',
}

function formatDateRange(start: string, end: string): string {
  const fmt = new Intl.DateTimeFormat('fr-FR', { day: 'numeric', month: 'short' })
  const startFmt = fmt.format(new Date(start))
  const endFmt = fmt.format(new Date(end))
  if (start === end) return startFmt
  return `${startFmt} — ${endFmt}`
}

function CategoryPill({ label }: { label: string }) {
  return <span className={styles.categoryPill}>{label}</span>
}

function EventRow({ event }: { event: SaleEvent }) {
  const isPast = event.daysUntil < 0 && !event.isActive
  const icon = TYPE_ICON[event.type] ?? '📅'

  return (
    <div
      className={`${styles.eventRow} ${isPast ? styles.eventRowPast : ''} ${event.isActive ? styles.eventRowActive : ''}`}
    >
      <span className={styles.eventIcon}>{icon}</span>
      <div className={styles.eventBody}>
        <div className={styles.eventTop}>
          <span className={styles.eventName}>{event.name}</span>
          <span className={styles.discountBadge}>{event.discountRange}</span>
        </div>
        <div className={styles.eventDate}>{formatDateRange(event.startDate, event.endDate)}</div>
        {event.tip && !isPast && (
          <div className={styles.eventTip}>{event.tip}</div>
        )}
        {event.categories.length > 0 && (
          <div className={styles.categoryRow}>
            {event.categories.map((cat) => (
              <CategoryPill key={cat} label={cat} />
            ))}
          </div>
        )}
      </div>
      {!isPast && (
        <div className={styles.countdown}>
          {event.isActive ? (
            <span className={styles.countdownActive}>EN COURS</span>
          ) : (
            <span className={styles.countdownDays}>J-{event.daysUntil}</span>
          )}
        </div>
      )}
    </div>
  )
}

function ActiveBanner({ event }: { event: SaleEvent }) {
  return (
    <div className={styles.activeBanner}>
      <span className={styles.activePulse} />
      <span className={styles.activeName}>{TYPE_ICON[event.type] ?? '📅'} {event.name}</span>
      <span className={styles.activeDiscount}>{event.discountRange} de réduction</span>
    </div>
  )
}

function NextEventBanner({ event }: { event: SaleEvent }) {
  return (
    <div className={styles.nextBanner}>
      <span className={styles.nextLabel}>PROCHAIN ÉVÉNEMENT</span>
      <span className={styles.nextName}>{event.name}</span>
      <span className={styles.nextCountdown}>dans {event.daysUntil} jour{event.daysUntil > 1 ? 's' : ''}</span>
    </div>
  )
}

export function SaleCalendarWidget() {
  const { events, nextEvent, activeEvent, isLoading } = useSaleCalendar()

  if (isLoading) {
    return (
      <div className={styles.panel}>
        <div className={styles.header}>
          <span className={styles.headerTitle}>📅 CALENDRIER DES SOLDES</span>
        </div>
        <div className={styles.skeletonBody}>
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className={styles.skeletonRow} />
          ))}
        </div>
      </div>
    )
  }

  if (events.length === 0) return null

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <span className={styles.headerTitle}>📅 CALENDRIER DES SOLDES</span>
        <span className={styles.headerSub}>Marché français 2026</span>
      </div>

      {activeEvent && <ActiveBanner event={activeEvent} />}
      {!activeEvent && nextEvent && <NextEventBanner event={nextEvent} />}

      <div className={styles.timeline}>
        {events.map((event) => (
          <EventRow key={event.id} event={event} />
        ))}
      </div>
    </div>
  )
}
