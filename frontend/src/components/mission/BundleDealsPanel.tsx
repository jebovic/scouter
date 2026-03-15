import { useBundleDeals } from '../../hooks/useBundleDeals'
import type { BundleDeal } from '../../api/bundledetector'
import styles from './BundleDealsPanel.module.css'

interface Props {
  missionId: string
}

const CONFIDENCE_LABEL: Record<string, string> = {
  high: '🟢 Élevée',
  medium: '🟡 Moyenne',
  low: '🔴 Faible',
}

const CONFIDENCE_CLASS: Record<string, string> = {
  high: styles.confidenceHigh,
  medium: styles.confidenceMedium,
  low: styles.confidenceLow,
}

const CARD_CLASS: Record<string, string> = {
  high: styles.high,
  medium: styles.medium,
  low: styles.low,
}

function fmt(n: number): string {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency: 'EUR' }).format(n)
}

function BundleCard({ bundle }: { bundle: BundleDeal }) {
  const cardClass = `${styles.bundleCard} ${CARD_CLASS[bundle.confidence] ?? ''}`
  const badgeClass = `${styles.confidenceBadge} ${CONFIDENCE_CLASS[bundle.confidence] ?? ''}`

  return (
    <div className={cardClass}>
      <div className={styles.bundleHeader}>
        <span className={styles.bundleName}>{bundle.bundleName}</span>
        <span className={badgeClass}>{CONFIDENCE_LABEL[bundle.confidence]}</span>
      </div>

      <div className={styles.items}>
        {bundle.items.map((item) => (
          <div key={item.itemId} className={styles.itemRow}>
            <span className={styles.itemName}>{item.itemName}</span>
            <span className={styles.itemPrice}>{fmt(item.price)}</span>
          </div>
        ))}
      </div>

      <div className={styles.savings}>
        <span className={styles.savingsTotal}>
          Total bundle : <strong>{fmt(bundle.totalPrice)}</strong>
        </span>
        <span className={styles.savingsAmount}>
          Économie estimée : {fmt(bundle.estimatedSave)} ({bundle.savePercent.toFixed(0)}%)
        </span>
      </div>

      <p className={styles.tip}>{bundle.tip}</p>

      <span className={styles.categoryBadge}>{bundle.category}</span>
    </div>
  )
}

export function BundleDealsPanel({ missionId }: Props) {
  const { bundleResponse, isLoading } = useBundleDeals(missionId)

  if (isLoading) return null

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <span className={styles.title}>📦 Opportunités de Bundle</span>
        {bundleResponse && bundleResponse.totalPotentialSave > 0 && (
          <span className={styles.totalSave}>
            Économie potentielle : {fmt(bundleResponse.totalPotentialSave)}
          </span>
        )}
      </div>

      {!bundleResponse || bundleResponse.bundles.length === 0 ? (
        <p className={styles.empty}>
          Aucun bundle détecté — ajoutez plus d'articles pour découvrir des synergies
        </p>
      ) : (
        <>
          <p className={styles.empty}>{bundleResponse.message}</p>
          <div className={styles.grid}>
            {bundleResponse.bundles.map((bundle, idx) => (
              <BundleCard key={`${bundle.items[0]?.itemId ?? idx}-${bundle.items[1]?.itemId ?? idx}`} bundle={bundle} />
            ))}
          </div>
        </>
      )}
    </div>
  )
}
