import { useQueries } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { useMissions } from '../hooks/useMission'
import { useSettings } from '../hooks/useSettings'
import { getPurchaseRecord } from '../api/purchase'
import { StarRating } from '../components/scouter/StarRating'
import { formatCurrencyLocale, formatDate } from '../utils/format'
import styles from './HistoryPage.module.css'

export default function HistoryPage() {
  const { t } = useTranslation()
  const { data: settings } = useSettings()
  const locale = settings?.locale ?? 'fr-FR'
  const currency = settings?.currency ?? 'EUR'

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
        <h1 className={styles.heading}>{t('purchase.history')}</h1>
        <div className={styles.empty}>
          <p className={styles.emptyIcon}>📦</p>
          <p className={styles.emptyTitle}>{t('purchase.noHistory')}</p>
          <p className={styles.emptyDesc}>{t('purchase.noHistoryDesc')}</p>
          <Link to="/" className={styles.cta}>{t('purchase.goToDashboard')}</Link>
        </div>
      </main>
    )
  }

  return (
    <main className={styles.page}>
      <h1 className={styles.heading}>{t('purchase.history')}</h1>
      <div className={styles.grid}>
        {doneMissions.map((mission, idx) => {
          const purchase = purchaseQueries[idx]?.data
          const savings = mission.budget && purchase ? mission.budget - purchase.finalPrice : null
          const missionCurrency = mission.currency ?? currency
          const missionLocale = mission.locale ?? locale
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
                    <span className={styles.price}>{formatCurrencyLocale(purchase.finalPrice, missionLocale, missionCurrency)}</span>
                    {savings !== null && (
                      <span className={`${styles.savings} ${savings >= 0 ? styles.positive : styles.negative}`}>
                        {savings >= 0
                          ? `−${formatCurrencyLocale(savings, missionLocale, missionCurrency)}`
                          : `+${formatCurrencyLocale(Math.abs(savings), missionLocale, missionCurrency)}`}
                        {' '}{t('purchase.vsBudget')}
                      </span>
                    )}
                  </div>
                  <p className={styles.meta}>
                    {purchase.merchant} · {formatDate(purchase.purchasedAt, missionLocale)}
                  </p>
                  {purchase.satisfaction && (
                    <StarRating value={purchase.satisfaction} readOnly size={16} />
                  )}
                  {purchase.review && <p className={styles.review}>{purchase.review}</p>}
                </div>
              ) : (
                <p className={styles.noPurchase}>{t('purchase.noRecord')}</p>
              )}
            </div>
          )
        })}
      </div>
    </main>
  )
}
