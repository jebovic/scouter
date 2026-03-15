import { useSalesCalendar } from '../../hooks/useSalesCalendar'
import { daysUntil } from '../../utils/frenchSalesCalendar'
import type { SaleEvent, SaleEventType } from '../../utils/frenchSalesCalendar'
import styles from './SalesCalendar.module.css'

interface SalesCalendarProps {
  category?: string
}

const TYPE_LABELS: Record<SaleEventType, string> = {
  soldes: 'Soldes',
  promo: 'Promo',
  flash: 'Flash',
  tech: 'Tech',
  seasonal: 'Saison',
}

function formatDateRange(startDate: string, endDate: string): string {
  const fmt = (d: string) =>
    new Date(d).toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' })
  if (startDate === endDate) return fmt(startDate)
  return `${fmt(startDate)} — ${fmt(endDate)}`
}

function StatusChip({ startDate }: { startDate: string }) {
  const days = daysUntil(startDate)
  if (days <= 0) {
    return <span className={`${styles.chip} ${styles.chipActive}`}>En cours</span>
  }
  if (days === 1) {
    return <span className={`${styles.chip} ${styles.chipSoon}`}>Demain</span>
  }
  return (
    <span className={`${styles.chip} ${styles.chipUpcoming}`}>
      dans {days}j
    </span>
  )
}

function EventCard({ event, highlight = false }: { event: SaleEvent; highlight?: boolean }) {
  return (
    <div className={`${styles.eventCard} ${styles[`type_${event.type}`]} ${highlight ? styles.highlight : ''}`}>
      <div className={styles.eventHeader}>
        <span className={styles.typeBadge}>{TYPE_LABELS[event.type]}</span>
        <StatusChip startDate={event.startDate} />
      </div>
      <div className={styles.eventName}>{event.name}</div>
      <div className={styles.eventDate}>{formatDateRange(event.startDate, event.endDate)}</div>
      <div className={styles.discountBadge}>{event.discountRange}</div>
    </div>
  )
}

export function SalesCalendar({ category = '' }: SalesCalendarProps) {
  const { upcomingEvents, nextBestEvent } = useSalesCalendar(category)

  return (
    <div className={styles.root}>
      <div className={styles.titleRow}>
        <span className={styles.flag}>🇫🇷</span>
        <h3 className={styles.title}>Calendrier des Bonnes Affaires</h3>
      </div>

      {/* Best time to buy callout */}
      {category && nextBestEvent && (
        <div className={styles.callout}>
          <div className={styles.calloutLabel}>Meilleur moment pour acheter</div>
          <div className={styles.calloutCategory}>
            Catégorie : <strong>{category}</strong>
          </div>
          <EventCard event={nextBestEvent} highlight />
        </div>
      )}

      {/* Upcoming events list */}
      {upcomingEvents.length === 0 ? (
        <p className={styles.empty}>Aucune promotion dans les 90 prochains jours.</p>
      ) : (
        <div className={styles.eventList}>
          {upcomingEvents.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      )}
    </div>
  )
}
