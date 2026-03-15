import { usePriceStats } from '../../hooks/usePriceStats'
import { formatCurrency } from '../../utils/format'
import styles from './PriceAnalyticsPanel.module.css'

interface PriceAnalyticsPanelProps {
  missionId: string
  itemId: string
  currency?: string
}

function volatilityLevel(avg: number, stdDev: number): 'low' | 'medium' | 'high' {
  if (avg === 0) return 'low'
  const cv = stdDev / avg // coefficient of variation
  if (cv < 0.05) return 'low'
  if (cv < 0.15) return 'medium'
  return 'high'
}

function trendArrow(slope: number): string {
  if (slope > 0.01) return '↑'
  if (slope < -0.01) return '↓'
  return '→'
}

function trendClass(slope: number): string {
  if (slope > 0.01) return 'rising'
  if (slope < -0.01) return 'falling'
  return 'stable'
}

function drop7dClass(drop: number): string {
  if (drop < -0.001) return 'falling'
  if (drop > 0.001) return 'rising'
  return 'stable'
}

export function PriceAnalyticsPanel({ missionId, itemId, currency = 'USD' }: PriceAnalyticsPanelProps) {
  const { stats, isLoading } = usePriceStats(missionId, itemId)
  const fmt = (n: number) => formatCurrency(n, currency)

  if (isLoading) {
    return (
      <div className={styles.panel}>
        <div className={styles.title}>Price Analytics</div>
        <div className={styles.noData}>Loading…</div>
      </div>
    )
  }

  if (!stats || stats.dataPoints === 0) {
    return (
      <div className={styles.panel}>
        <div className={styles.title}>Price Analytics</div>
        <div className={styles.noData}>No price history available yet.</div>
      </div>
    )
  }

  const volLevel = volatilityLevel(stats.average, stats.volatility)
  const arrow = trendArrow(stats.trendSlope)
  const trendCls = trendClass(stats.trendSlope)
  const dropCls = drop7dClass(stats.priceDrop7d)
  const drop7dSign = stats.priceDrop7d > 0.001 ? '+' : ''

  return (
    <div className={styles.panel}>
      <div className={styles.title}>Price Analytics ({stats.dataPoints} data points)</div>

      <div className={styles.grid}>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Min</span>
          <span className={`${styles.statValue} ${styles.falling}`}>{fmt(stats.min)}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Average</span>
          <span className={styles.statValue}>{fmt(stats.average)}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Max</span>
          <span className={`${styles.statValue} ${styles.rising}`}>{fmt(stats.max)}</span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Trend</span>
          <span className={`${styles.statValue} ${styles[trendCls]}`}>
            {arrow} {trendCls}
          </span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>vs 7 days ago</span>
          <span className={`${styles.statValue} ${styles[dropCls]}`}>
            {stats.priceDrop7d === 0 ? '—' : `${drop7dSign}${fmt(stats.priceDrop7d)}`}
          </span>
        </div>
        <div className={styles.stat}>
          <span className={styles.statLabel}>Std Dev</span>
          <span className={styles.statValue}>{fmt(stats.volatility)}</span>
        </div>
      </div>

      <div className={styles.volatilityRow}>
        <span className={styles.volatilityLabel}>Volatility:</span>
        <span className={`${styles.volatilityBadge} ${styles[volLevel]}`}>{volLevel.toUpperCase()}</span>
      </div>
    </div>
  )
}
