import { usePurchaseAdvice } from '../../hooks/usePurchaseAdvice'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { PurchaseAdvice } from '../../api/purchaseadvisor'
import styles from './PurchaseAdvisorCard.module.css'

interface PurchaseAdvisorCardProps {
  itemId: string
  itemName: string
}

const DECISION_LABELS: Record<PurchaseAdvice['decision'], string> = {
  buy_now: 'Acheter maintenant',
  wait: 'Attendre',
  skip: 'Ignorer',
  negotiate: 'Négocier',
}

const DECISION_CLASS: Record<PurchaseAdvice['decision'], string> = {
  buy_now: styles.decisionBuyNow,
  wait: styles.decisionWait,
  skip: styles.decisionSkip,
  negotiate: styles.decisionNegotiate,
}

function LoadingDots() {
  return (
    <span className={styles.loadingDots} aria-label="Analyse en cours">
      <span />
      <span />
      <span />
    </span>
  )
}

function ConfidenceBar({ value }: { value: number }) {
  return (
    <div className={styles.confidenceBar} role="progressbar" aria-valuenow={value} aria-valuemin={0} aria-valuemax={100}>
      <div className={styles.confidenceFill} style={{ '--pct': `${value}%` } as React.CSSProperties} />
      <span className={styles.confidenceLabel}>{value}% de confiance</span>
    </div>
  )
}

export function PurchaseAdvisorCard({ itemId, itemName }: PurchaseAdvisorCardProps) {
  const { advice, isPending, getAdvice } = usePurchaseAdvice(itemId)
  const { fmt } = useFormatCurrency()

  return (
    <div className={styles.card}>
      {!advice && !isPending && (
        <button
          className={styles.triggerBtn}
          onClick={getAdvice}
          aria-label={`Demander un conseil IA pour ${itemName}`}
        >
          Demander à l&apos;IA
        </button>
      )}

      {isPending && (
        <div className={styles.loadingRow}>
          <LoadingDots />
          <span className={styles.loadingText}>Analyse en cours...</span>
        </div>
      )}

      {advice && (
        <div className={styles.result}>
          <div className={styles.decisionRow}>
            <span className={`${styles.decisionBadge} ${DECISION_CLASS[advice.decision]}`}>
              {DECISION_LABELS[advice.decision]}
            </span>
          </div>

          <ConfidenceBar value={advice.confidence} />

          <p className={styles.reasoning}>{advice.reasoning}</p>

          {advice.bestTimeToAct && (
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Meilleur moment</span>
              <span className={styles.metaValue}>{advice.bestTimeToAct}</span>
            </div>
          )}

          {advice.estimatedSavingsByWaiting > 0 && (
            <div className={styles.metaRow}>
              <span className={styles.metaLabel}>Économies potentielles</span>
              <span className={styles.metaValue}>
                {fmt(advice.estimatedSavingsByWaiting)}
              </span>
            </div>
          )}

          {advice.alternativeSuggestion && (
            <div className={styles.alternative}>
              <span className={styles.alternativeLabel}>Alternative</span>
              <p className={styles.alternativeText}>{advice.alternativeSuggestion}</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
