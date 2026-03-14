import type { PriceTrend } from '../../types'
import styles from './TrendBadge.module.css'

interface TrendBadgeProps {
  trend: PriceTrend
}

const TREND_CONFIG: Record<PriceTrend, { label: string; icon: string }> = {
  dropping: { label: 'dropping', icon: '↓' },
  stable: { label: 'stable', icon: '→' },
  rising: { label: 'rising', icon: '↑' },
}

export function TrendBadge({ trend }: TrendBadgeProps) {
  const { label, icon } = TREND_CONFIG[trend]
  return (
    <span className={`${styles.badge} ${styles[trend]}`}>
      <span className={styles.icon} aria-hidden="true">{icon}</span>
      {label}
    </span>
  )
}
