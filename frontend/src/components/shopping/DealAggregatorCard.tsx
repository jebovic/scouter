import { useDealAggregator } from '../../hooks/useDealAggregator'
import type { Deal } from '../../api/dealaggregator'
import styles from './DealAggregatorCard.module.css'

interface DealAggregatorCardProps {
  missionId: string
  itemId: string
  currency?: string
}

function DealRow({ deal }: { deal: Deal }) {
  const fmtPrice = (n: number) =>
    n.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })

  return (
    <div className={styles.deal}>
      <span className={styles.source}>{deal.source}</span>
      <span className={styles.dealTitle} title={deal.title}>{deal.title}</span>
      <span className={styles.discount}>-{deal.discount}%</span>
      <span className={styles.price}>{fmtPrice(deal.dealPrice)}</span>
      {deal.hot && <span className={styles.hot} aria-label="Hot deal">🔥</span>}
      <span className={styles.votes}>{deal.votes} votes</span>
      <a
        href={deal.url}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.link}
        aria-label={`Voir le deal ${deal.source}`}
      >
        Voir →
      </a>
    </div>
  )
}

export function DealAggregatorCard({ missionId, itemId }: DealAggregatorCardProps) {
  const { data, isLoading } = useDealAggregator(missionId, itemId)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.skeleton} aria-busy="true" aria-label="Chargement des bons plans" />
      </div>
    )
  }

  if (!data || data.deals.length === 0) {
    return (
      <div className={styles.card}>
        <p className={styles.empty}>Aucun bon plan trouvé</p>
      </div>
    )
  }

  const fmtPrice = (n: number) =>
    n.toLocaleString('fr-FR', { style: 'currency', currency: 'EUR' })

  return (
    <div className={styles.card}>
      <p className={styles.title}>🏷️ Bons Plans</p>

      {data.bestDeal && (
        <div className={styles.bestDeal}>
          <div className={styles.bestDealLabel}>
            Meilleur deal {data.bestDeal.hot ? '🔥' : ''}
          </div>
          <div className={styles.deal}>
            <span className={styles.source}>{data.bestDeal.source}</span>
            <span className={styles.dealTitle} title={data.bestDeal.title}>
              {data.bestDeal.title}
            </span>
            <span className={styles.discount}>-{data.bestDeal.discount}%</span>
            <span className={styles.price}>{fmtPrice(data.bestDeal.dealPrice)}</span>
            <span className={styles.votes}>{data.bestDeal.votes} votes</span>
            <a
              href={data.bestDeal.url}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.link}
              aria-label={`Voir le meilleur deal ${data.bestDeal.source}`}
            >
              Voir →
            </a>
          </div>
        </div>
      )}

      {data.deals.map((deal, i) => (
        <DealRow key={`${deal.source}-${i}`} deal={deal} />
      ))}
    </div>
  )
}
