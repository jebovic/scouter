import { useState } from 'react'
import { usePromoFeed } from '../../hooks/useDealFeed'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { PromoOffer } from '../../api/dealfeed'
import styles from './PromoFeedPanel.module.css'

interface PromoFeedPanelProps {
  query: string
  category?: string
}

function ExpiresLabel({ expiresAt }: { expiresAt: string | null | undefined }) {
  if (!expiresAt) return null
  const date = new Date(expiresAt)
  const now = Date.now()
  const diff = date.getTime() - now
  if (diff <= 0) return <span className={styles.expiredBadge}>Expirée</span>
  const hours = Math.floor(diff / (1000 * 60 * 60))
  const days = Math.floor(hours / 24)
  const label = days > 0 ? `Expire dans ${days}j` : `Expire dans ${hours}h`
  return <span className={styles.expiresBadge}>{label}</span>
}

function OfferCard({ offer }: { offer: PromoOffer }) {
  const { fmt } = useFormatCurrency()
  return (
    <article className={`${styles.card} ${offer.isFlash ? styles.cardFlash : ''}`}>
      <div className={styles.cardHeader}>
        <span className={styles.retailer}>{offer.retailer}</span>
        {offer.isFlash && (
          <span className={styles.flashBadge}>
            <span className={styles.flashIcon}>⚡</span>
            Vente flash
          </span>
        )}
        <span className={styles.discountBadge}>-{offer.discount}%</span>
      </div>

      <h3 className={styles.offerTitle}>{offer.title}</h3>

      <div className={styles.priceRow}>
        <span className={styles.originalPrice}>{fmt(offer.originalPrice)}</span>
        <span className={styles.promoPrice}>{fmt(offer.promoPrice)}</span>
      </div>

      <div className={styles.cardFooter}>
        <ExpiresLabel expiresAt={offer.expiresAt} />
        {offer.url && (
          <a
            href={`https://${offer.url}`}
            target="_blank"
            rel="noopener noreferrer"
            className={styles.link}
          >
            Voir l'offre →
          </a>
        )}
      </div>
    </article>
  )
}

function SkeletonCard() {
  return (
    <div className={styles.skeletonCard}>
      <div className={styles.skeletonLine} style={{ width: '40%' } as React.CSSProperties} />
      <div className={styles.skeletonLine} style={{ width: '75%' } as React.CSSProperties} />
      <div className={styles.skeletonPriceLine} />
    </div>
  )
}

export function PromoFeedPanel({ query, category }: PromoFeedPanelProps) {
  const [triggered, setTriggered] = useState(false)
  const { offers, isLoading, error } = usePromoFeed(query, category, triggered)

  const canSearch = query.length >= 2

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h2 className={styles.panelTitle}>Bons plans du moment</h2>
        {!triggered && (
          <button
            className={styles.triggerBtn}
            onClick={() => setTriggered(true)}
            disabled={!canSearch}
            title={canSearch ? undefined : 'Saisissez au moins 2 caractères'}
          >
            Trouver des promos
          </button>
        )}
      </div>

      {triggered && isLoading && (
        <div className={styles.grid}>
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      )}

      {triggered && error && (
        <div className={styles.errorBox}>
          Impossible de charger les promos. Veuillez réessayer.
        </div>
      )}

      {triggered && !isLoading && !error && offers.length === 0 && (
        <div className={styles.emptyBox}>
          Aucune promotion trouvée pour cette recherche.
        </div>
      )}

      {triggered && !isLoading && offers.length > 0 && (
        <>
          <div className={styles.grid}>
            {offers.map((offer, i) => (
              <OfferCard key={`${offer.retailer}-${i}`} offer={offer} />
            ))}
          </div>
          <p className={styles.disclaimer}>
            Offres générées par IA — vérifier avant d'acheter
          </p>
        </>
      )}
    </section>
  )
}
