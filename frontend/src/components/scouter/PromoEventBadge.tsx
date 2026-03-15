import type { PromoEvent } from '../../utils/frenchPromoCalendar'
import styles from './PromoEventBadge.module.css'

interface PromoEventBadgeProps {
  event: PromoEvent
  compact?: boolean
  onClick?: (event: PromoEvent) => void
}

export function PromoEventBadge({ event, compact = false, onClick }: PromoEventBadgeProps) {
  return (
    <button
      className={styles.badge}
      onClick={() => onClick?.(event)}
      title={event.description}
      style={{
        '--badge-color': event.color,
      } as React.CSSProperties}
    >
      <span className={styles.icon}>{event.icon}</span>
      <span className={styles.name}>{event.name}</span>
      {!compact && <span className={styles.discount}>{event.discount}</span>}
    </button>
  )
}
