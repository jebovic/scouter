import type { NegotiationScript } from '../../api/negotiation'
import styles from './NegotiationCoach.module.css'

interface NegotiationCoachProps {
  script: NegotiationScript
  onClose: () => void
}

function formatCurrency(value: number): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(value)
}

export function NegotiationCoach({ script, onClose }: NegotiationCoachProps) {
  const {
    openingOffer,
    walkAwayPrice,
    suggestedDiscount,
    talkingPoints,
    counterOfferScript,
    confidenceScore,
  } = script

  const confPct = Math.min(100, Math.max(0, Math.round(confidenceScore)))

  return (
    <div className={styles.panel} role="region" aria-label="Negotiation Coach">
      <div className={styles.header}>
        <h4 className={styles.title}>NEGOTIATION COACH</h4>
        <button className={styles.closeBtn} onClick={onClose} aria-label="Close negotiation coach">
          ×
        </button>
      </div>

      {/* Price grid */}
      <div className={styles.priceGrid}>
        <div className={styles.priceCell}>
          <span className={styles.priceCellLabel}>Opening Offer</span>
          <span className={`${styles.priceCellValue} ${styles.opening}`}>
            {formatCurrency(openingOffer)}
          </span>
        </div>
        <div className={styles.priceCell}>
          <span className={styles.priceCellLabel}>Walk-Away</span>
          <span className={`${styles.priceCellValue} ${styles.walkaway}`}>
            {formatCurrency(walkAwayPrice)}
          </span>
        </div>
        <div className={styles.priceCell}>
          <span className={styles.priceCellLabel}>Target Discount</span>
          <span className={styles.discountBadge}>{suggestedDiscount}% OFF</span>
        </div>
      </div>

      {/* Confidence bar */}
      <div className={styles.confidenceRow}>
        <div className={styles.confidenceLabel}>
          <span>Confidence</span>
          <span className={styles.confidenceValue}>{confPct}%</span>
        </div>
        <div className={styles.confidenceTrack}>
          <div
            className={styles.confidenceFill}
            style={{ '--conf-width': `${confPct}%` } as React.CSSProperties}
          />
        </div>
      </div>

      {/* Talking points */}
      {talkingPoints.length > 0 && (
        <div className={styles.section}>
          <div className={styles.sectionLabel}>Talking Points</div>
          <ol className={styles.talkingList}>
            {talkingPoints.map((point, i) => (
              <li key={i} className={styles.talkingItem}>
                <span className={styles.talkingNum}>{i + 1}.</span>
                <span>{point}</span>
              </li>
            ))}
          </ol>
        </div>
      )}

      {/* Counter-offer script */}
      {counterOfferScript.length > 0 && (
        <div className={styles.section}>
          <div className={styles.sectionLabel}>Counter-Offer Script</div>
          <div className={styles.scriptList}>
            {counterOfferScript.map((line, i) => (
              <div key={i} className={styles.scriptBubble}>
                <span className={styles.scriptStep}>S{i + 1}</span>
                <span className={styles.scriptText}>{line}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
