import { useMemo } from 'react'
import { getUpcomingEvents, isEventActive } from '../../utils/frenchPromoCalendar'
import type { PromoEvent } from '../../utils/frenchPromoCalendar'
import { PromoEventBadge } from './PromoEventBadge'
import styles from './UpcomingPromoStrip.module.css'

interface UpcomingPromoStripProps {
  onEventClick?: (event: PromoEvent) => void
}

export function UpcomingPromoStrip({ onEventClick }: UpcomingPromoStripProps) {
  const upcomingEvents = useMemo(() => {
    return getUpcomingEvents(new Date())
  }, [])

  // Hide strip if no upcoming events
  if (upcomingEvents.length === 0) {
    return null
  }

  return (
    <section className={styles.strip} role="region" aria-label="Promotions à venir">
      <div className={styles.header}>
        <h2 className={styles.title}>Promotions à venir</h2>
      </div>
      <div className={styles.scroll}>
        <div className={styles.content}>
          {upcomingEvents.map((event) => {
            const isActive = isEventActive(event, new Date())
            return (
              <div
                key={event.id}
                className={`${styles.item} ${isActive ? styles.itemActive : ''}`}
              >
                <PromoEventBadge
                  event={event}
                  compact={false}
                  onClick={onEventClick}
                />
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}
