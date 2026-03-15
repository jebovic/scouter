import { useState } from 'react'
import { useSubstitutes, useFetchSubstitutes } from '../../hooks/useSubstitutes'
import type { Substitute } from '../../api/substitutes'
import styles from './SubstituteSuggestions.module.css'

interface Props {
  optionId: string
}

function formatPrice(price: number, currency: string): string {
  return new Intl.NumberFormat('fr-FR', { style: 'currency', currency }).format(price)
}

function googleShoppingUrl(name: string): string {
  return `https://www.google.com/search?tbm=shop&q=${encodeURIComponent(name + ' France')}`
}

function LoadingSkeleton() {
  return (
    <div className={styles.skeleton}>
      {[1, 2, 3].map((i) => (
        <div
          key={i}
          className={styles.skeletonLine}
          style={{ width: i === 1 ? '80%' : i === 2 ? '65%' : '50%' }}
        />
      ))}
    </div>
  )
}

function SubstituteItem({ sub }: { sub: Substitute }) {
  return (
    <div className={styles.item}>
      {/* Header: name + price + savings badge */}
      <div className={styles.itemHeader}>
        <div className={styles.nameGroup}>
          <span className={styles.name}>{sub.name}</span>
          <span className={styles.brand}>{sub.brand}</span>
        </div>
        <div className={styles.priceGroup}>
          <span className={styles.price}>{formatPrice(sub.price, sub.currency)}</span>
          {sub.savingsPercent != null && sub.savingsPercent > 0 && (
            <span className={styles.savingsBadge}>-{Math.round(sub.savingsPercent)}%</span>
          )}
        </div>
      </div>

      {/* Pros / Cons */}
      {((sub.pros && sub.pros.length > 0) || (sub.cons && sub.cons.length > 0)) && (
        <div className={styles.prosConsRow}>
          {sub.pros && sub.pros.length > 0 && (
            <div className={styles.prosList}>
              <span className={styles.prosTitle}>Avantages</span>
              {sub.pros.map((p, i) => (
                <span key={i} className={styles.prosBullet}>
                  {p}
                </span>
              ))}
            </div>
          )}
          {sub.cons && sub.cons.length > 0 && (
            <div className={styles.consList}>
              <span className={styles.consTitle}>Inconvénients</span>
              {sub.cons.map((c, i) => (
                <span key={i} className={styles.consBullet}>
                  {c}
                </span>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Why Consider */}
      {sub.whyConsider && (
        <p className={styles.whyConsider}>{sub.whyConsider}</p>
      )}

      {/* Google Shopping link */}
      <a
        href={googleShoppingUrl(sub.name)}
        target="_blank"
        rel="noopener noreferrer"
        className={styles.googleLink}
      >
        Rechercher sur Google Shopping →
      </a>
    </div>
  )
}

function averageSavings(subs: Substitute[]): number | null {
  const withSavings = subs.filter((s) => s.savingsPercent != null && s.savingsPercent > 0)
  if (withSavings.length === 0) return null
  const total = withSavings.reduce((acc, s) => acc + (s.savingsPercent ?? 0), 0)
  return Math.round(total / withSavings.length)
}

function averageSavingsAmount(originalPrice: number, subs: Substitute[]): number | null {
  const avg = averageSavings(subs)
  if (avg == null || originalPrice <= 0) return null
  return Math.round((originalPrice * avg) / 100)
}

export function SubstituteSuggestions({ optionId }: Props) {
  const [open, setOpen] = useState(false)

  const { data, isLoading } = useSubstitutes(open ? optionId : undefined)
  const { fetch, isPending } = useFetchSubstitutes(optionId)

  const handleOpen = () => {
    setOpen(true)
    fetch()
  }

  if (!open) {
    return (
      <div className={styles.wrapper}>
        <button className={styles.triggerBtn} onClick={handleOpen}>
          Alternatives moins chères
        </button>
      </div>
    )
  }

  const loading = isLoading || isPending

  return (
    <div className={styles.wrapper}>
      <div className={styles.content}>
        <div className={styles.sectionTitle}>Alternatives moins chères</div>

        {loading && <LoadingSkeleton />}

        {!loading && data && data.substitutes.length === 0 && (
          <p className={styles.empty}>Aucune alternative trouvée</p>
        )}

        {!loading && data && data.substitutes.length > 0 && (
          <>
            <div className={styles.list}>
              {data.substitutes.map((sub, idx) => (
                <SubstituteItem key={`${sub.name}-${idx}`} sub={sub} />
              ))}
            </div>

            {/* Average savings context footer */}
            {(() => {
              const originalPrice = data.originalPrice ?? 0
              const avgAmount = averageSavingsAmount(originalPrice, data.substitutes)
              const avgPct = averageSavings(data.substitutes)
              if (avgPct == null) return null
              return (
                <div className={styles.savingsContext}>
                  {avgAmount != null && avgAmount > 0
                    ? `Ces alternatives vous économiseraient ${formatPrice(avgAmount, 'EUR')} en moyenne (−${avgPct}% vs ${data.productName}).`
                    : `Ces alternatives représentent en moyenne ${avgPct}% d'économies.`}
                </div>
              )
            })()}
          </>
        )}
      </div>
    </div>
  )
}
