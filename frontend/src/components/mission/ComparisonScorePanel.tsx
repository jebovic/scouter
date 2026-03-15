import { useState } from 'react'
import { useComparisonScore } from '../../hooks/useComparisonScore'
import { LoadingPulse } from '../scouter'
import type { ItemScore } from '../../api/comparisonscore'
import styles from './ComparisonScorePanel.module.css'

interface Props {
  missionId: string
  currency: string
}

function formatCurrency(value: number, currency: string): string {
  const locale = currency === 'EUR' ? 'fr-FR' : 'en-US'
  return new Intl.NumberFormat(locale, { style: 'currency', currency }).format(value)
}

function getRankClass(rank: number): string {
  switch (rank) {
    case 1:
      return styles.rank1
    case 2:
      return styles.rank2
    case 3:
      return styles.rank3
    default:
      return styles.rankDefault
  }
}

function getScoreBarClass(score: number): string {
  if (score >= 80) {
    return styles.scoreBarFillGreen
  } else if (score >= 60) {
    return styles.scoreBarFillOrange
  }
  return styles.scoreBarFillRed
}

function ItemScoreRow({ item, currency }: { item: ItemScore; currency: string }) {
  return (
    <div className={styles.itemRow}>
      <div className={`${styles.rank} ${getRankClass(item.rank)}`}>{item.rank}</div>
      <div className={styles.itemName}>{item.name}</div>
      <div className={styles.priceValue}>{formatCurrency(item.price, currency)}</div>
      <div className={styles.scoreBarContainer}>
        <div className={styles.scoreBar}>
          <div
            className={`${styles.scoreBarFill} ${getScoreBarClass(item.compositeScore)}`}
            style={{ width: `${Math.min(100, Math.max(0, item.compositeScore))}%` } as React.CSSProperties}
          />
        </div>
        <div className={styles.scoreLabel}>{item.compositeScore.toFixed(0)}/100</div>
      </div>
      <div className={styles.verdict}>{item.verdict}</div>
    </div>
  )
}

export function ComparisonScorePanel({ missionId, currency }: Props) {
  const { data: report, isLoading } = useComparisonScore(missionId)
  const [isExpanded, setIsExpanded] = useState(true)

  if (isLoading) {
    return <LoadingPulse label="Calcul des scores..." />
  }

  if (!report || report.itemScores.length === 0) {
    return (
      <div className={styles.panel}>
        <div className={styles.header}>
          <span className={styles.title}>📊 Score de Comparaison</span>
        </div>
        <p className={styles.empty}>{report?.summary || 'Aucun article à comparer'}</p>
      </div>
    )
  }

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <span className={styles.title}>📊 Score de Comparaison</span>
        {report.summary && <span className={styles.summaryBadge}>{report.summary}</span>}
      </div>

      <button
        className={styles.toggle}
        onClick={() => setIsExpanded(!isExpanded)}
        aria-expanded={isExpanded}
      >
        <span className={`${styles.toggleIcon} ${isExpanded ? styles.open : ''}`}>▼</span>
        {isExpanded ? 'Masquer les scores' : 'Voir les scores'}
      </button>

      {isExpanded && (
        <div className={styles.list}>
          {report.itemScores.map((item) => (
            <ItemScoreRow key={item.itemId} item={item} currency={currency} />
          ))}
        </div>
      )}
    </div>
  )
}
