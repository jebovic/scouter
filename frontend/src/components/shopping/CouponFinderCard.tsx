import { useState } from 'react'
import { useCoupons } from '../../hooks/useCoupons'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { Coupon, CashbackOffer } from '../../api/couponfinder'
import styles from './CouponFinderCard.module.css'

interface CouponFinderCardProps {
  missionId: string
  itemId: string
}

const PLATFORM_EMOJI: Record<string, string> = {
  iGraal: '🟢',
  Poulpeo: '🔵',
  Widilo: '🟠',
  Rakuten: '🔴',
}

function confidenceClass(confidence: string): string {
  if (confidence === 'vérifié') return styles.confVerifie
  if (confidence === 'signalé') return styles.confSignale
  return styles.confExpire
}

function CouponRow({ coupon }: { coupon: Coupon }) {
  const [copied, setCopied] = useState(false)

  function handleCopy() {
    navigator.clipboard.writeText(coupon.code).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    })
  }

  return (
    <div className={styles.couponRow}>
      <div className={styles.codeWrap}>
        <span className={styles.code}>{coupon.code}</span>
        <button
          className={styles.copyBtn}
          onClick={handleCopy}
          aria-label={`Copier le code ${coupon.code}`}
          title="Copier le code"
        >
          {copied ? '✓' : 'Copier'}
        </button>
      </div>
      <div className={styles.couponInfo}>
        <div className={styles.couponDesc}>{coupon.description}</div>
        <div className={styles.couponMeta}>
          <span className={styles.source}>{coupon.source}</span>
          <span aria-hidden="true">·</span>
          <span className={styles.expires}>⏱ {coupon.expiresIn}</span>
          <span
            className={`${styles.confidenceBadge} ${confidenceClass(coupon.confidence)}`}
          >
            {coupon.confidence}
          </span>
        </div>
      </div>
      <div className={styles.discountPill} aria-label={`-${coupon.discountPct}%`}>
        -{coupon.discountPct}%
      </div>
    </div>
  )
}

function CashbackRow({ offer, fmt }: { offer: CashbackOffer; fmt: (n: number) => string }) {
  const emoji = PLATFORM_EMOJI[offer.platform] ?? '💰'
  return (
    <div className={styles.cashbackRow}>
      <span className={styles.platformEmoji} aria-hidden="true">{emoji}</span>
      <div className={styles.cashbackInfo}>
        <div className={styles.platformName}>{offer.platform}</div>
        <div className={styles.cashbackDesc}>
          {offer.description}
          {offer.minPurchase > 0 && (
            <> · min. {fmt(offer.minPurchase)}</>
          )}
        </div>
      </div>
      <div className={styles.cashbackPct} aria-label={`${offer.cashbackPct}% cashback`}>
        {offer.cashbackPct}%
      </div>
    </div>
  )
}

export function CouponFinderCard({ missionId, itemId }: CouponFinderCardProps) {
  const [enabled, setEnabled] = useState(false)
  const [open, setOpen] = useState(false)

  const { data, isFetching } = useCoupons(missionId, itemId, enabled)
  const { fmt } = useFormatCurrency()

  function handleToggle() {
    setEnabled(true)
    setOpen((v) => !v)
  }

  return (
    <div className={styles.wrapper}>
      <button
        className={styles.triggerBtn}
        onClick={handleToggle}
        aria-expanded={open}
        aria-label="Codes promo et cashback"
        title="Trouver des codes promo et offres cashback"
      >
        {isFetching ? <span className={styles.spinner} aria-hidden="true" /> : '🏷️ Promos'}
      </button>

      {open && (
        <div className={styles.panel} role="region" aria-label="Codes promo et cashback">
          {isFetching && !data ? (
            <p className={styles.loading}>Recherche en cours...</p>
          ) : data ? (
            <>
              <div className={styles.header}>
                <span aria-hidden="true">🏷️</span>
                Codes Promo &amp; Cashback
              </div>

              {data.coupons.length > 0 && (
                <>
                  <div className={styles.sectionTitle}>Codes promo</div>
                  <div className={styles.couponList}>
                    {data.coupons.map((c) => (
                      <CouponRow key={c.code} coupon={c} />
                    ))}
                  </div>
                </>
              )}

              {data.cashbackOffers.length > 0 && (
                <>
                  <div className={styles.sectionTitle}>Cashback disponible</div>
                  <div className={styles.cashbackList}>
                    {data.cashbackOffers.map((cb) => (
                      <CashbackRow key={cb.platform} offer={cb} fmt={fmt} />
                    ))}
                  </div>
                </>
              )}

              {data.totalPotentialSaving > 0 && (
                <div className={styles.savingsBanner} role="status">
                  <span aria-hidden="true">💰</span>
                  Économie potentielle totale&nbsp;:&nbsp;
                  <span className={styles.savingsAmt}>
                    {fmt(data.totalPotentialSaving)}
                  </span>
                </div>
              )}

              <p className={styles.disclaimer}>
                Codes et taux indicatifs — vérifiez leur validité avant achat.
              </p>

              <button
                className={styles.closeBtn}
                onClick={() => setOpen(false)}
                aria-label="Fermer les promos"
              >
                Fermer
              </button>
            </>
          ) : null}
        </div>
      )}
    </div>
  )
}
