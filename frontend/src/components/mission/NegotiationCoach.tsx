import { useTranslation } from 'react-i18next'
import type { NegotiationScript } from '../../api/negotiation'
import { useSettings } from '../../hooks/useSettings'
import { formatCurrencyLocale } from '../../utils/format'
import styles from './NegotiationCoach.module.css'

interface NegotiationCoachProps {
  script: NegotiationScript
  onClose: () => void
}

export function NegotiationCoach({ script, onClose }: NegotiationCoachProps) {
  const { t } = useTranslation()
  const { data: settings } = useSettings()
  const locale = settings?.locale ?? 'fr-FR'
  const currency = settings?.currency ?? 'EUR'

  const {
    openingOffer,
    walkAwayPrice,
    suggestedDiscount,
    talkingPoints,
    counterOfferScript,
    confidenceScore,
  } = script

  const confPct = Math.min(100, Math.max(0, Math.round(confidenceScore)))
  const fmt = (v: number) => formatCurrencyLocale(v, locale, currency)

  return (
    <div className={styles.panel} role="region" aria-label={t('negotiation.title')}>
      <div className={styles.header}>
        <h4 className={styles.title}>{t('negotiation.title').toUpperCase()}</h4>
        <button className={styles.closeBtn} onClick={onClose} aria-label={t('common.close')}>
          ×
        </button>
      </div>

      {/* Price grid */}
      <div className={styles.priceGrid}>
        <div className={styles.priceCell}>
          <span className={styles.priceCellLabel}>{t('negotiation.openingOffer')}</span>
          <span className={`${styles.priceCellValue} ${styles.opening}`}>
            {fmt(openingOffer)}
          </span>
        </div>
        <div className={styles.priceCell}>
          <span className={styles.priceCellLabel}>{t('negotiation.walkAwayPrice')}</span>
          <span className={`${styles.priceCellValue} ${styles.walkaway}`}>
            {fmt(walkAwayPrice)}
          </span>
        </div>
        <div className={styles.priceCell}>
          <span className={styles.priceCellLabel}>{t('negotiation.targetDiscount')}</span>
          <span className={styles.discountBadge}>{suggestedDiscount}% {t('negotiation.off')}</span>
        </div>
      </div>

      {/* Confidence bar */}
      <div className={styles.confidenceRow}>
        <div className={styles.confidenceLabel}>
          <span>{t('negotiation.confidence')}</span>
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
          <div className={styles.sectionLabel}>{t('negotiation.talkingPoints')}</div>
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
          <div className={styles.sectionLabel}>{t('negotiation.counterScript')}</div>
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
