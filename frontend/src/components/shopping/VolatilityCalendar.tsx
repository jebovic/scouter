import { useVolatilityCalendar } from '../../hooks/useVolatilityCalendar'
import { formatCurrency } from '../../utils/format'
import styles from './VolatilityCalendar.module.css'

interface VolatilityCalendarProps {
  missionId: string
  itemId: string
  currency?: string
}

/**
 * Interpolates between green (stable) → gold (mid) → coral (volatile)
 * based on a 0–1 intensity value.
 */
function intensityToColor(intensity: number): string {
  // green: #00d68f, gold: #ffd93d, coral: #ff3d71
  if (intensity <= 0.5) {
    const t = intensity * 2
    const r = Math.round(0 + t * (255 - 0))
    const g = Math.round(214 + t * (217 - 214))
    const b = Math.round(143 + t * (61 - 143))
    return `rgb(${r},${g},${b})`
  } else {
    const t = (intensity - 0.5) * 2
    const r = Math.round(255 + t * (255 - 255))
    const g = Math.round(217 + t * (61 - 217))
    const b = Math.round(61 + t * (113 - 61))
    return `rgb(${r},${g},${b})`
  }
}

export function VolatilityCalendar({ missionId, itemId, currency = 'EUR' }: VolatilityCalendarProps) {
  const { calendar, isLoading } = useVolatilityCalendar(missionId, itemId)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Volatilité des prix</div>
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)', fontStyle: 'italic' }}>
          Analyse en cours…
        </div>
      </div>
    )
  }

  if (!calendar || calendar.weeks.length === 0) {
    return null
  }

  return (
    <div className={styles.card}>
      <div className={styles.title}>Volatilité des prix — 90 jours</div>

      {/* Heatmap grid */}
      <div className={styles.grid}>
        {calendar.weeks.map((week) => (
          <div
            key={`${week.year}-${week.week}`}
            className={styles.cell}
            style={
              { backgroundColor: intensityToColor(week.intensity), opacity: 0.25 + week.intensity * 0.75 } as React.CSSProperties
            }
          >
            <div className={styles.tooltip}>
              <div className={styles.tooltipLabel}>{week.label} ({week.year})</div>
              <div className={styles.tooltipRange}>
                {formatCurrency(week.minPrice, currency)} – {formatCurrency(week.maxPrice, currency)}
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Legend */}
      <div className={styles.legend}>
        <span className={styles.legendLabel}>Stable</span>
        <div className={styles.legendGradient} />
        <span className={styles.legendLabel}>Volatile</span>
      </div>

      {/* Most volatile week */}
      {calendar.mostVolatileWeek && (
        <div className={styles.mostVolatile}>
          Semaine la plus volatile :{' '}
          <span className={styles.mostVolatileHighlight}>
            {calendar.mostVolatileWeek.label}
          </span>{' '}
          ({formatCurrency(calendar.mostVolatileWeek.minPrice, currency)} –{' '}
          {formatCurrency(calendar.mostVolatileWeek.maxPrice, currency)})
        </div>
      )}

      {/* Recommendation */}
      {calendar.recommendation && (
        <div className={styles.recommendation}>{calendar.recommendation}</div>
      )}
    </div>
  )
}
