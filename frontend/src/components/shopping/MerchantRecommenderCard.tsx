import { useState } from 'react'
import { useMerchantRecommender } from '../../hooks/useMerchantRecommender'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { MerchantScore } from '../../api/merchantrecommender'
import styles from './MerchantRecommenderCard.module.css'

interface MerchantRecommenderCardProps {
  missionId: string
  itemId: string
  itemName?: string
}

function MerchantRow({ score, fmt }: { score: MerchantScore; fmt: (n: number) => string }) {

  const getScoreClass = (s: number) => {
    if (s >= 80) return 'excellent'
    if (s >= 60) return 'good'
    if (s >= 40) return 'average'
    return 'poor'
  }

  return (
    <div className={styles.merchantRow}>
      <div className={styles.merchantName}>{score.merchant}</div>
      <div className={styles.scoreContainer}>
        <div className={styles.scoreBar}>
          <div
            className={`${styles.scoreFill} ${styles[getScoreClass(score.score)]}`}
            style={{ width: `${Math.min(score.score, 100)}%` }}
            aria-valuenow={Math.round(score.score)}
            aria-valuemin={0}
            aria-valuemax={100}
            role="progressbar"
          />
        </div>
        <div className={styles.scoreLabel}>{Math.round(score.score)}</div>
      </div>
      <div className={styles.priceRecommendation}>
        <div className={styles.price}>{fmt(score.avgPrice)}</div>
        <div className={styles.recommendation}>{score.recommendation}</div>
      </div>
    </div>
  )
}

export function MerchantRecommenderCard({
  missionId,
  itemId,
}: MerchantRecommenderCardProps) {
  const [showAll, setShowAll] = useState(false)
  const { data, isLoading } = useMerchantRecommender(missionId, itemId)
  const { fmt } = useFormatCurrency()

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.skeleton} aria-busy="true" aria-label="Chargement des recommandations de marchands" />
      </div>
    )
  }

  if (!data || data.rankings.length === 0) {
    return (
      <div className={styles.card}>
        <p className={styles.empty}>Aucune recommandation de marchand disponible</p>
      </div>
    )
  }

  const displayCount = showAll ? data.rankings.length : 5
  const displayedMerchants = data.rankings.slice(0, displayCount)
  const hasMore = data.rankings.length > 5

  return (
    <div className={styles.card}>
      <p className={styles.title}>🏪 Recommandations de Marchands</p>

      <div className={styles.merchantList}>
        {displayedMerchants.map((score) => (
          <MerchantRow key={score.merchant} score={score} fmt={fmt} />
        ))}
      </div>

      {hasMore && !showAll && (
        <button
          className={styles.showMore}
          onClick={() => setShowAll(true)}
          aria-label={`Afficher les ${data.rankings.length - 5} marchands supplémentaires`}
        >
          Afficher {data.rankings.length - 5} marchands supplémentaires
        </button>
      )}

      {showAll && hasMore && (
        <button
          className={styles.showMore}
          onClick={() => setShowAll(false)}
          aria-label="Afficher moins de marchands"
        >
          Afficher moins
        </button>
      )}
    </div>
  )
}
