import { useState } from 'react'
import { Link } from 'react-router-dom'
import { usePriceDigest } from '../../hooks/usePriceDigest'
import { useFormatCurrency } from '../../hooks/useFormatCurrency'
import type { PriceChange } from '../../api/pricedigest'
import styles from './PriceDigestWidget.module.css'

const MAX_VISIBLE = 5

function minutesAgo(isoDate: string): number {
  return Math.floor((Date.now() - new Date(isoDate).getTime()) / 60_000)
}

interface ChangeRowProps {
  change: PriceChange
}

function ChangeRow({ change }: ChangeRowProps) {
  const { fmt } = useFormatCurrency()
  const pctLabel = `${change.isDropping ? '-' : '+'}${Math.abs(change.changePct).toFixed(1)}%`

  return (
    <div className={styles.dropRow} data-rising={String(!change.isDropping)}>
      <Link
        to={`/missions/${change.missionSlug}`}
        className={styles.itemLink}
        title={`${change.itemName} — ${change.missionName}`}
      >
        {change.itemName}
      </Link>
      <span className={styles.merchant}>{change.merchant}</span>
      <div className={styles.prices}>
        <span className={styles.oldPrice}>{fmt(change.oldPrice)}</span>
        <span className={styles.arrow2}>→</span>
        <span className={styles.newPrice} data-dropping={String(change.isDropping)}>
          {fmt(change.newPrice)}
        </span>
      </div>
      <span className={styles.changePct} data-dropping={String(change.isDropping)}>
        {pctLabel}
      </span>
    </div>
  )
}

export function PriceDigestWidget() {
  const { digest, isLoading } = usePriceDigest()
  const [expanded, setExpanded] = useState(false)

  if (isLoading || !digest) return null

  const hasChanges = digest.priceDrops.length > 0 || digest.priceRises.length > 0

  const visibleDrops = expanded ? digest.priceDrops.slice(0, MAX_VISIBLE) : []
  const ago = minutesAgo(digest.generatedAt)
  const agoLabel = ago < 1 ? "à l'instant" : `il y a ${ago} min`

  return (
    <div className={styles.widget} role="region" aria-label="Digest des variations de prix">
      <div className={styles.header}>
        <div className={styles.titleRow}>
          <span className={styles.title}>Variations de prix 24h</span>
          {hasChanges && (
            <div className={styles.counters}>
              {digest.priceDrops.length > 0 && (
                <span className={styles.counter} data-type="drops" title="Baisses de prix">
                  <span className={styles.arrow}>↓</span>
                  {digest.priceDrops.length}
                </span>
              )}
              {digest.priceRises.length > 0 && (
                <span className={styles.counter} data-type="rises" title="Hausses de prix">
                  <span className={styles.arrow}>↑</span>
                  {digest.priceRises.length}
                </span>
              )}
            </div>
          )}
        </div>

        {hasChanges && (
          <button
            className={styles.toggleBtn}
            onClick={() => setExpanded((v) => !v)}
            aria-expanded={expanded}
          >
            {expanded ? 'Masquer' : 'Voir les baisses'}
          </button>
        )}
      </div>

      {!hasChanges && (
        <p className={styles.stable}>Aucun changement de prix détecté</p>
      )}

      {expanded && visibleDrops.length > 0 && (
        <div className={styles.dropList}>
          {visibleDrops.map((change) => (
            <ChangeRow key={change.itemId} change={change} />
          ))}
        </div>
      )}

      <p className={styles.timestamp}>Mis à jour {agoLabel}</p>
    </div>
  )
}
