import { useQueries } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useMissions } from '../hooks/useMission'
import { getPurchaseRecord } from '../api/purchase'
import { StarRating } from '../components/scouter/StarRating'
import styles from './HistoryPage.module.css'

function formatCurrency(amount: number, currency?: string) {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: currency || 'EUR' }).format(amount)
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('fr-FR', { year: 'numeric', month: 'long', day: 'numeric' })
}

export default function HistoryPage() {
  const { missions = [] } = useMissions()
  const doneMissions = missions.filter((m) => m.phase === 'done')

  const purchaseQueries = useQueries({
    queries: doneMissions.map((m) => ({
      queryKey: ['purchase', m.id],
      queryFn: () => getPurchaseRecord(m.id),
    })),
  })

  if (doneMissions.length === 0) {
    return (
      <main className={styles.page}>
        <h1 className={styles.heading}>Purchase History</h1>
        <div className={styles.empty}>
          <p className={styles.emptyIcon}>📦</p>
          <p className={styles.emptyTitle}>No completed missions yet</p>
          <p className={styles.emptyDesc}>Finish a research mission and record a purchase to see it here.</p>
          <Link to="/" className={styles.cta}>Go to Dashboard</Link>
        </div>
      </main>
    )
  }

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>Purchase History</h1>
      <div className={styles.grid}>
        {doneMissions.map((mission, idx) => {
          const purchase = purchaseQueries[idx]?.data
          const savings = mission.budget && purchase ? mission.budget - purchase.finalPrice : null
          return (
            <div key={mission.id} className={styles.card}>
              <div className={styles.cardHeader}>
                <Link to={`/missions/${mission.slug}`} className={styles.missionName}>
                  {mission.name}
                </Link>
                {mission.category && <span className={styles.category}>{mission.category}</span>}
              </div>
              {purchase ? (
                <div className={styles.purchaseDetails}>
                  <div className={styles.priceRow}>
                    <span className={styles.price}>{formatCurrency(purchase.finalPrice)}</span>
                    {savings !== null && (
                      <span className={`${styles.savings} ${savings >= 0 ? styles.positive : styles.negative}`}>
                        {savings >= 0 ? `−${formatCurrency(savings)}` : `+${formatCurrency(Math.abs(savings))}`}
                        {' vs budget'}
                      </span>
                    )}
                  </div>
                  <p className={styles.meta}>
                    {purchase.merchant} · {formatDate(purchase.purchasedAt)}
                  </p>
                  {purchase.satisfaction && (
                    <StarRating value={purchase.satisfaction} readOnly size={16} />
                  )}
                  {purchase.review && <p className={styles.review}>{purchase.review}</p>}
                </div>
              ) : (
                <p className={styles.noPurchase}>No purchase recorded</p>
              )}
            </div>
          )
        })}
      </div>
    </main>
  )
}
