import { useState } from 'react'
import { useWishlistShare } from '../../hooks/useWishlistShare'
import type { WishlistItem } from '../../api/wishlistshare'
import styles from './WishlistShareCard.module.css'

interface Props {
  missionId: string
  currency?: string
}

function statusClass(status: string): string {
  switch (status) {
    case 'buy': return styles.statusBuy
    case 'flash-sale': return styles.statusFlashSale
    case 'preorder': return styles.statusPreorder
    case 'defer': return styles.statusDefer
    case 'watch': return styles.statusWatch
    case 'crisis': return styles.statusCrisis
    default: return styles.statusDefault
  }
}

function formatPrice(price: number, currency: string): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(price)
}

function ItemRow({ item, currency }: { item: WishlistItem; currency: string }) {
  return (
    <div className={styles.item}>
      <span className={styles.itemName}>{item.name}</span>
      {item.merchant && (
        <span className={styles.itemMerchant}>{item.merchant}</span>
      )}
      <span className={styles.itemPrice}>{formatPrice(item.price, currency)}</span>
      {item.status && (
        <span className={`${styles.statusBadge} ${statusClass(item.status)}`}>
          {item.status}
        </span>
      )}
    </div>
  )
}

export function WishlistShareCard({ missionId, currency = 'EUR' }: Props) {
  const { data, isLoading, isError } = useWishlistShare(missionId)
  const [copied, setCopied] = useState(false)

  if (isLoading) {
    return (
      <div className={styles.card}>
        <div className={styles.empty}>Loading wishlist…</div>
      </div>
    )
  }

  if (isError || !data) {
    return (
      <div className={styles.card}>
        <div className={styles.empty}>Unable to load wishlist card.</div>
      </div>
    )
  }

  const shareUrl = data.shareUrl || `${window.location.origin}${window.location.pathname}?share=wishlist`
  const effectiveCurrency = data.currency || currency

  function handleCopyLink() {
    navigator.clipboard.writeText(shareUrl).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }

  async function handleShare() {
    if (navigator.share) {
      try {
        await navigator.share({
          title: `${data!.missionName} — Wishlist`,
          text: `Check out my wishlist for ${data!.missionName}`,
          url: shareUrl,
        })
      } catch {
        // User cancelled or browser rejected — fall back to clipboard
        handleCopyLink()
      }
    } else {
      handleCopyLink()
    }
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div>
          <div className={styles.title}>{data.missionName}</div>
          <div className={styles.budget}>
            Budget: {formatPrice(data.totalBudget, effectiveCurrency)}
            {' · '}
            {data.items.length} item{data.items.length !== 1 ? 's' : ''}
          </div>
        </div>
        <div className={styles.actions}>
          {copied ? (
            <span className={styles.copied}>Copied!</span>
          ) : (
            <button className={styles.copyBtn} onClick={handleCopyLink}>
              Copy Link
            </button>
          )}
          <button className={styles.shareBtn} onClick={handleShare}>
            Share
          </button>
        </div>
      </div>

      {data.items.length === 0 ? (
        <div className={styles.empty}>No items in this wishlist yet.</div>
      ) : (
        <div className={styles.itemList}>
          {data.items.map((item, i) => (
            <ItemRow key={i} item={item} currency={effectiveCurrency} />
          ))}
        </div>
      )}
    </div>
  )
}
