import type { DealScore } from '../../types'
import styles from './DealScoreBadge.module.css'

interface DealScoreBadgeProps {
  score: DealScore
}

export function DealScoreBadge({ score }: DealScoreBadgeProps) {
  const pct = score.pctBelowAvg
  const isGood = pct > 0
  const label = pct === 0
    ? 'at avg'
    : isGood
      ? `${pct.toFixed(1)}% below avg`
      : `${Math.abs(pct).toFixed(1)}% above avg`

  return (
    <span className={`${styles.badge} ${isGood ? styles.good : styles.bad}`} title={`Historical avg: ${score.historicalAvg.toFixed(2)}`}>
      {label}
    </span>
  )
}
