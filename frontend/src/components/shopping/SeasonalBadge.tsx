import { useState } from 'react'
import { useSeasonal } from '../../hooks/useSeasonal'
import styles from './SeasonalBadge.module.css'

interface SeasonalBadgeProps {
  itemName: string
  category?: string
}

function urgencyClass(nextEventDays: number): string {
  if (nextEventDays < 30) return styles.soon
  if (nextEventDays < 90) return styles.near
  return styles.later
}

function abbreviateMonths(months: string[]): string {
  return months
    .map((m) => m.slice(0, 3))
    .join(' · ')
}

export function SeasonalBadge({ itemName, category = '' }: SeasonalBadgeProps) {
  const [expanded, setExpanded] = useState(false)
  const { data, isLoading, isError } = useSeasonal(itemName, category)

  if (isLoading) return <span className={styles.loading}>🗓️ …</span>
  if (isError || !data) return null

  const cls = urgencyClass(data.nextEventDays)
  const monthsLabel = abbreviateMonths(data.bestMonths)

  return (
    <div className={styles.wrap}>
      <button
        className={`${styles.pill} ${cls}`}
        onClick={() => setExpanded((v) => !v)}
        title={`Meilleur moment : ${data.bestMonths.join(', ')} — ${data.avgDiscountPct}% de remise attendue`}
        aria-expanded={expanded}
      >
        🗓️ {monthsLabel}
      </button>
      {expanded && (
        <div className={styles.details}>
          <div className={styles.detailRow}>
            <span className={styles.label}>Remise moy.</span>
            <span className={styles.value}>{data.avgDiscountPct.toFixed(0)} %</span>
          </div>
          {data.nextEventDays > 0 && (
            <div className={styles.detailRow}>
              <span className={styles.label}>Prochaine opportunité</span>
              <span className={styles.value}>dans {data.nextEventDays} j</span>
            </div>
          )}
          {data.events.length > 0 && (
            <div>
              <span className={styles.label}>Événements</span>
              <ul className={styles.events}>
                {data.events.map((ev, i) => (
                  <li key={i}>{ev}</li>
                ))}
              </ul>
            </div>
          )}
          {data.rationale && (
            <p className={styles.rationale}>{data.rationale}</p>
          )}
        </div>
      )}
    </div>
  )
}
