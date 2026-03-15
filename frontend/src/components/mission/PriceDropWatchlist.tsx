import { useState } from 'react'
import { usePriceDropWatch } from '../../hooks/usePriceDropWatch'
import { LoadingPulse } from '../scouter'
import styles from './PriceDropWatchlist.module.css'

interface PriceDropWatchlistProps {
  missionId: string
}

export function PriceDropWatchlist({ missionId }: PriceDropWatchlistProps) {
  const { data, isLoading } = usePriceDropWatch(missionId)
  const [isExpanded, setIsExpanded] = useState(true)

  if (isLoading) {
    return <LoadingPulse label="Chargement de la surveillance..." />
  }

  if (!data || data.watchItems.length === 0) {
    return (
      <div className={styles.card}>
        <div className={styles.header}>
          <h3 className={styles.title}>📉 Surveillance Baisses de Prix</h3>
        </div>
        <p className={styles.emptyText}>Aucun signal de baisse détecté actuellement.</p>
      </div>
    )
  }

  return (
    <div className={styles.card}>
      {/* Header */}
      <div className={styles.header}>
        <button
          className={styles.expandBtn}
          onClick={() => setIsExpanded(!isExpanded)}
          aria-label={isExpanded ? 'Réduire' : 'Développer'}
        >
          {isExpanded ? '▼' : '▶'}
        </button>
        <h3 className={styles.title}>📉 Surveillance Baisses de Prix</h3>
      </div>

      {/* Summary */}
      <p className={styles.summary}>{data.summary}</p>

      {/* Expanded content */}
      {isExpanded && (
        <div className={styles.items}>
          {data.watchItems.map((item) => (
            <div key={item.itemId} className={styles.itemRow}>
              {/* Item name and price */}
              <div className={styles.itemInfo}>
                <div className={styles.itemName}>{item.name}</div>
                <div className={styles.itemPrice}>
                  {new Intl.NumberFormat('fr-FR', {
                    style: 'currency',
                    currency: 'EUR',
                  }).format(item.currentPrice)}
                </div>
              </div>

              {/* Probability bar */}
              <div className={styles.barContainer}>
                <div
                  className={`${styles.probabilityBar} ${
                    item.dropProbability >= 70
                      ? styles.highRisk
                      : item.dropProbability >= 40
                        ? styles.mediumRisk
                        : styles.lowRisk
                  }`}
                  style={{
                    '--probability': `${item.dropProbability}%`,
                  } as React.CSSProperties}
                />
                <span className={styles.probabilityLabel}>
                  {Math.round(item.dropProbability)}%
                </span>
              </div>

              {/* Signal badge */}
              <div
                className={`${styles.signalBadge} ${
                  item.signal === 'high_volatility'
                    ? styles.volatility
                    : item.signal === 'recent_spike'
                      ? styles.spike
                      : item.signal === 'no_history'
                        ? styles.noHistory
                        : styles.stable
                }`}
              >
                {item.signal === 'high_volatility'
                  ? 'Volatilité'
                  : item.signal === 'recent_spike'
                    ? 'Pic'
                    : item.signal === 'no_history'
                      ? 'Nouveau'
                      : 'Stable'}
              </div>

              {/* Expected drop */}
              <div className={styles.expectedDrop}>
                {Math.round(item.expectedDrop)}%
              </div>

              {/* Reasoning */}
              <div className={styles.reasoning}>{item.reasoning}</div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
