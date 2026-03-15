import { usePriceInsights } from '../../hooks/usePriceInsights'
import { formatCurrency } from '../../utils/format'
import styles from './PriceInsightsCard.module.css'

interface PriceInsightsCardProps {
  missionId: string
  itemId: string
  currency?: string
}

export function PriceInsightsCard({ missionId, itemId, currency = 'EUR' }: PriceInsightsCardProps) {
  const { insights, isLoading } = usePriceInsights(missionId, itemId)
  const fmt = (n: number) => formatCurrency(n, currency)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Meilleur moment pour acheter</div>
        <div className={styles.noData}>Chargement…</div>
      </div>
    )
  }

  if (!insights || insights.dataPoints < 3) {
    return (
      <div className={styles.card}>
        <div className={styles.header}>
          <div className={styles.title}>Meilleur moment pour acheter</div>
          {insights && <span className={`${styles.confidenceBadge} ${styles.low}`}>faible confiance</span>}
        </div>
        <div className={styles.noData}>
          {insights?.buyRecommendation ?? 'Pas assez de données pour analyser les tendances de prix.'}
        </div>
      </div>
    )
  }

  const maxPrice = Math.max(...insights.priceByDay.map((d) => d.avgPrice), 0.01)

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.title}>Meilleur moment pour acheter ({insights.dataPoints} données)</div>
        <span className={`${styles.confidenceBadge} ${styles[insights.confidence]}`}>
          confiance {insights.confidence === 'high' ? 'haute' : insights.confidence === 'medium' ? 'moyenne' : 'faible'}
        </span>
      </div>

      <div className={styles.highlights}>
        <div className={styles.highlight}>
          <span className={styles.highlightLabel}>Meilleur jour</span>
          <span className={`${styles.highlightValue} ${styles.best}`}>{insights.bestDayOfWeek}</span>
        </div>
        <div className={styles.highlight}>
          <span className={styles.highlightLabel}>Moins bon jour</span>
          <span className={`${styles.highlightValue} ${styles.worst}`}>{insights.worstDayOfWeek}</span>
        </div>
        <div className={styles.highlight}>
          <span className={styles.highlightLabel}>Période du mois</span>
          <span className={`${styles.highlightValue} ${styles.best}`}>{insights.bestTimeOfMonth} du mois</span>
        </div>
      </div>

      {insights.priceByDay.length > 0 && (
        <div className={styles.chartSection}>
          <div className={styles.chartLabel}>Prix moyen par jour de la semaine</div>
          <div className={styles.bars}>
            {insights.priceByDay.map((day) => {
              const widthPct = maxPrice > 0 ? (day.avgPrice / maxPrice) * 100 : 0
              const isBest = day.day === insights.bestDayOfWeek
              const isWorst = day.day === insights.worstDayOfWeek
              const fillClass = isBest ? styles.best : isWorst ? styles.worst : ''
              return (
                <div key={day.day} className={styles.barRow}>
                  <span className={styles.barDay}>{day.day}</span>
                  <div className={styles.barTrack}>
                    <div
                      className={`${styles.barFill} ${fillClass}`}
                      style={{ width: `${widthPct}%` }}
                    />
                  </div>
                  <span className={styles.barPrice}>{fmt(day.avgPrice)}</span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      <div className={styles.recommendation}>{insights.buyRecommendation}</div>
    </div>
  )
}
