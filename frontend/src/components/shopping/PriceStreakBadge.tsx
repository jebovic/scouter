import { usePriceStreak } from '../../hooks/usePriceStreak'
import styles from './PriceStreakBadge.module.css'

interface PriceStreakBadgeProps {
  missionId: string
  itemId: string
}

export function PriceStreakBadge({ missionId, itemId }: PriceStreakBadgeProps) {
  const { streak, isLoading } = usePriceStreak(missionId, itemId)

  if (isLoading || !streak || streak.currentStreak < 2) return null

  const icon = streak.streakType === 'dropping' ? '📉' : '📈'
  const sign = streak.streakType === 'dropping' ? '-' : '+'
  const label = `${icon} ${sign}${streak.currentStreak} jours`

  const urgencyClass = {
    low: styles.urgencyLow,
    medium: styles.urgencyMedium,
    high: styles.urgencyHigh,
    urgent: styles.urgencyUrgent,
  }[streak.urgency]

  const pctAbs = Math.abs(streak.totalDropPct)
  const pctSign = streak.totalDropPct < 0 ? '-' : '+'
  const pctLabel = `${pctSign}${pctAbs.toFixed(1)}%`

  return (
    <div className={styles.wrapper}>
      <span className={`${styles.badge} ${urgencyClass}`}>
        {label}
      </span>
      <div className={styles.tooltip}>
        <div className={styles.tooltipPct}>{pctLabel} sur la période</div>
        <div className={styles.tooltipRec}>{streak.recommendation}</div>
      </div>
    </div>
  )
}
