import { useState, useMemo } from 'react'
import { useRetailerRadar } from '../../hooks/useRetailerRadar'
import styles from './RetailerRadar.module.css'

interface RetailerRadarProps {
  missionSlug: string // actually a missionId, kept as missionSlug for API consistency
}

const MEDALS = ['🥇', '🥈', '🥉']
const MAX_ROWS_INITIAL = 8

export function RetailerRadar({ missionSlug }: RetailerRadarProps) {
  const { stats } = useRetailerRadar(missionSlug)
  const [showMore, setShowMore] = useState(false)

  // Use mission ID from slug if needed - for now, we use the slug as the missionId
  const displayStats = useMemo(() => {
    if (showMore) return stats
    return stats.slice(0, MAX_ROWS_INITIAL)
  }, [stats, showMore])

  if (stats.length === 0) {
    return (
      <div className={styles.root}>
        <div className={styles.header}>
          <span>🏪</span>
          <h3 className={styles.title}>Radar Revendeurs</h3>
        </div>
        <div className={styles.empty}>Aucun article trouvé</div>
      </div>
    )
  }

  const formatPrice = (price: number, currency: string) =>
    new Intl.NumberFormat('fr-FR', { style: 'currency', currency, minimumFractionDigits: 0 }).format(price)

  return (
    <div className={styles.root}>
      <div className={styles.header}>
        <span>🏪</span>
        <h3 className={styles.title}>Radar Revendeurs</h3>
      </div>

      <div className={styles.table}>
        {displayStats.map((stat, index) => (
          <div
            key={stat.itemId}
            className={`${styles.row} ${index === 0 ? styles.bestPrice : ''} ${!stat.inStock ? styles.outOfStock : ''}`}
          >
            <div className={styles.rank}>
              {MEDALS[index] ? (
                <span className={styles.rankMedal}>{MEDALS[index]}</span>
              ) : (
                <span className={styles.rankNumber}>{index + 1}.</span>
              )}
            </div>

            <div className={styles.merchant}>{stat.merchant}</div>

            <div className={styles.itemName}>
              {stat.itemName.length > 30 ? `${stat.itemName.substring(0, 27)}...` : stat.itemName}
            </div>

            <div className={styles.price}>{formatPrice(stat.bestPrice, stat.currency)}</div>

            {!stat.inStock && <div className={styles.stockBadge}>Rupture</div>}
          </div>
        ))}
      </div>

      {stats.length > MAX_ROWS_INITIAL && (
        <div className={styles.footer}>
          <button
            className={styles.toggleButton}
            onClick={() => setShowMore(!showMore)}
            aria-expanded={showMore}
          >
            {showMore ? '← Afficher moins' : `Voir plus (${stats.length - MAX_ROWS_INITIAL})`}
          </button>
        </div>
      )}
    </div>
  )
}
