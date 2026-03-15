import { usePriceAnnotations } from '../../hooks/usePriceAnnotations'
import type { AnnotationType } from '../../api/priceannotation'
import styles from './PriceAnnotationList.module.css'

const ICONS: Record<AnnotationType, string> = {
  new_low: '📉',
  rebound: '📈',
  spike_up: '⬆️',
  spike_down: '⬇️',
  stable: '➡️',
}

interface PriceAnnotationListProps {
  missionId: string
  itemId: string
  currency?: string
}

function formatDate(isoString: string): string {
  return new Date(isoString).toLocaleDateString('fr-FR', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

function formatPrice(price: number, currency: string): string {
  return new Intl.NumberFormat('fr-FR', {
    style: 'currency',
    currency,
    maximumFractionDigits: 0,
  }).format(price)
}

export function PriceAnnotationList({
  missionId,
  itemId,
  currency = 'EUR',
}: PriceAnnotationListProps) {
  const { annotations, isLoading } = usePriceAnnotations(missionId, itemId)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.title}>Annotations de prix</div>
        <div style={{ fontSize: '0.8rem', color: 'var(--text-dim)', fontStyle: 'italic' }}>
          Analyse en cours…
        </div>
      </div>
    )
  }

  if (!annotations || annotations.annotations.length === 0) {
    return null
  }

  return (
    <div className={styles.card}>
      <div className={styles.title}>Annotations de prix</div>
      <div className={styles.timeline} role="list" aria-label="Événements de prix">
        {annotations.annotations.map((a, i) => (
          <div key={i} className={styles.row} role="listitem">
            <span className={styles.icon} aria-hidden="true">
              {ICONS[a.type]}
            </span>
            <div className={styles.body}>
              <span className={`${styles.label} ${styles[a.type]}`}>{a.label}</span>
              <div className={styles.meta}>
                <span className={styles.date}>{formatDate(a.recordedAt)}</span>
                <span className={styles.price}>{formatPrice(a.price, currency)}</span>
              </div>
              <span className={styles.description}>{a.description}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
