import { useState } from 'react'
import { useLivePrices } from '../../hooks/useLivePrices'
import type { LivePriceListing } from '../../api/liveprices'
import styles from './LivePricesPanel.module.css'

// Categories that support live price lookup
const SUPPORTED_CATEGORIES = new Set([
  'food',
  'grocery',
  'groceries',
  'cooking',
  'custom',
])

interface LivePricesPanelProps {
  missionId: string
  optionId: string
  missionCategory: string
}

function formatPrice(price: number, currency: string): string {
  try {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(price)
  } catch {
    return `${currency} ${price.toFixed(2)}`
  }
}

interface ThumbnailProps {
  src: string
  alt: string
}

function Thumbnail({ src, alt }: ThumbnailProps) {
  const [errored, setErrored] = useState(false)

  if (!src || errored) {
    return <div className={styles.thumbnailFallback}>🏷</div>
  }

  return (
    <img
      src={src}
      alt={alt}
      className={styles.thumbnail}
      onError={() => setErrored(true)}
    />
  )
}

interface ListingRowProps {
  listing: LivePriceListing
  isCached: boolean
}

function ListingRow({ listing, isCached }: ListingRowProps) {
  const hasPrice = listing.price > 0

  return (
    <div className={styles.listing}>
      <Thumbnail src={listing.imageUrl} alt={listing.productName} />

      <div className={styles.listingBody}>
        <div className={styles.productName}>
          {listing.productName || listing.brand || 'Unknown product'}
        </div>
        <div className={styles.productMeta}>
          {[listing.brand, listing.stores, listing.condition]
            .filter(Boolean)
            .join(' · ')}
        </div>
      </div>

      <div className={styles.listingRight}>
        {hasPrice ? (
          <span className={styles.price}>
            {formatPrice(listing.price, listing.currency)}
          </span>
        ) : (
          <span className={styles.priceUnavailable}>N/A</span>
        )}

        <div className={styles.badges}>
          {isCached && <span className={styles.cachedBadge}>cached</span>}
          {listing.source && (
            <span className={styles.sourceBadge}>{listing.source}</span>
          )}
        </div>

        {listing.url && (
          <a
            href={listing.url}
            target="_blank"
            rel="noopener noreferrer"
            className={styles.listingLink}
          >
            View →
          </a>
        )}
      </div>
    </div>
  )
}

function SkeletonRows() {
  return (
    <>
      {[0, 1, 2].map((i) => (
        <div key={i} className={styles.skeletonRow}>
          <div className={styles.skeletonThumb} />
          <div className={styles.skeletonText}>
            <div className={styles.skeletonLine} />
            <div className={`${styles.skeletonLine} ${styles.skeletonShort}`} />
          </div>
        </div>
      ))}
    </>
  )
}

export function LivePricesPanel({ missionId, optionId, missionCategory }: LivePricesPanelProps) {
  const [isExpanded, setIsExpanded] = useState(false)

  const isSupported = SUPPORTED_CATEGORIES.has(missionCategory.toLowerCase())

  const { listings, isCached, isLoading, isError } = useLivePrices(
    missionId,
    optionId,
    isExpanded && isSupported,
  )

  if (!isSupported) {
    return null
  }

  return (
    <div className={styles.panel}>
      <button
        className={styles.header}
        onClick={() => setIsExpanded((prev) => !prev)}
        aria-expanded={isExpanded}
        type="button"
      >
        <div className={styles.headerLeft}>
          <span className={`${styles.chevron}${isExpanded ? ` ${styles.chevronOpen}` : ''}`}>
            ▶
          </span>
          <span className={styles.headerTitle}>Live Prices</span>
        </div>

        {isExpanded && listings.length > 0 && (
          <div className={styles.headerMeta}>
            <span className={styles.sourceBadge}>{listings.length} result{listings.length !== 1 ? 's' : ''}</span>
          </div>
        )}
      </button>

      {isExpanded && (
        <div className={styles.body}>
          {isLoading && <SkeletonRows />}

          {isError && !isLoading && (
            <div className={styles.errorState}>
              Price lookup unavailable — try again later.
            </div>
          )}

          {!isLoading && !isError && listings.length === 0 && (
            <div className={styles.emptyState}>No live prices found.</div>
          )}

          {!isLoading && !isError && listings.map((listing, i) => (
            <ListingRow
              key={`${listing.source}-${listing.url || i}`}
              listing={listing}
              isCached={isCached}
            />
          ))}
        </div>
      )}
    </div>
  )
}
