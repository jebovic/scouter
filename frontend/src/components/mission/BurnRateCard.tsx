import { useTranslation } from 'react-i18next'
import { useBurnRate } from '../../hooks/useBurnRate'
import styles from './BurnRateCard.module.css'

interface Props {
  missionId: string
  currency?: string
}

function formatCurrency(amount: number, currency = 'EUR'): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(amount)
}

function statusBadgeClass(status: string): string {
  switch (status) {
    case 'on_track':
      return styles.onTrack
    case 'at_risk':
      return styles.atRisk
    case 'over_budget':
      return styles.overBudget
    default:
      return styles.onTrack
  }
}

export function BurnRateCard({ missionId, currency = 'EUR' }: Props) {
  const { t } = useTranslation()
  const { data, isLoading, isError } = useBurnRate(missionId)

  const statusLabels: Record<string, string> = {
    on_track: t('burnRate.statusOnTrack'),
    at_risk: t('burnRate.statusAtRisk'),
    over_budget: t('burnRate.statusOverBudget'),
  }

  if (isLoading) {
    return (
      <section className={styles.card} aria-label={t('burnRate.ariaLabel')}>
        <div className={styles.header}>
          <span className={styles.title}>📈 {t('burnRate.title')}</span>
        </div>
        <div className={`${styles.skeleton} ${styles.skeletonLine}`} />
        <div className={styles.skeletonGrid}>
          {[1, 2, 3].map((i) => (
            <div key={i} className={`${styles.skeleton} ${styles.skeletonCell}`} />
          ))}
        </div>
      </section>
    )
  }

  if (isError || !data) {
    return (
      <section className={styles.card} aria-label={t('burnRate.ariaLabel')}>
        <div className={styles.header}>
          <span className={styles.title}>📈 {t('burnRate.title')}</span>
        </div>
        <p className={styles.errorState}>{t('burnRate.error')}</p>
      </section>
    )
  }

  const maxAmount = data.burnPoints.length > 0
    ? data.burnPoints[data.burnPoints.length - 1].amount
    : 1

  return (
    <section className={styles.card} aria-label={t('burnRate.ariaLabel')}>
      <div className={styles.header}>
        <span className={styles.title}>📈 {t('burnRate.title')}</span>
        <span className={`${styles.statusBadge} ${statusBadgeClass(data.status)}`}>
          {statusLabels[data.status] ?? data.status}
        </span>
      </div>

      <div className={styles.metrics}>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>{t('burnRate.metricDailyRate')}</span>
          <span className={styles.metricValue}>{formatCurrency(data.dailyRate, currency)}</span>
        </div>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>{t('burnRate.metricProjection')}</span>
          <span className={styles.metricValue}>{formatCurrency(data.projectedTotal, currency)}</span>
        </div>
        <div className={styles.metric}>
          <span className={styles.metricLabel}>{t('burnRate.metricDaysElapsed')}</span>
          <span className={styles.metricValue}>{data.daysElapsed}{t('burnRate.daysUnit')}</span>
        </div>
      </div>

      {data.burnPoints.length === 0 ? (
        <p className={styles.empty}>{t('burnRate.empty')}</p>
      ) : (
        <div className={styles.burnBars}>
          <span className={styles.burnBarsLabel}>{t('burnRate.cumulativeLabel')}</span>
          {data.burnPoints.map((pt) => {
            const pct = maxAmount > 0 ? Math.min((pt.amount / maxAmount) * 100, 100) : 0
            const isOver = pt.amount > data.totalBudget
            return (
              <div key={pt.date} className={styles.barItem}>
                <span className={styles.barDate}>{pt.date}</span>
                <div className={styles.barTrack}>
                  <div
                    className={styles.barFill}
                    style={{ '--pct': pct } as React.CSSProperties}
                    data-over={isOver ? 'true' : 'false'}
                  />
                </div>
                <span className={styles.barAmount}>{formatCurrency(pt.amount, currency)}</span>
              </div>
            )
          })}
        </div>
      )}
    </section>
  )
}
