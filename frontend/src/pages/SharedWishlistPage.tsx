import { useSearchParams } from 'react-router-dom'
import { decodeWishlist } from '../utils/wishlistShare'
import type { CompactWishlistItem } from '../utils/wishlistShare'
import styles from './SharedWishlistPage.module.css'

function formatPrice(price: number, currency: string): string {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(price)
  } catch {
    return `${price} ${currency}`
  }
}

function WishlistItemRow({ item }: { item: CompactWishlistItem }) {
  const nameContent = item.u ? (
    <a
      href={item.u}
      className={styles.itemName}
      target="_blank"
      rel="noopener noreferrer"
    >
      {item.n}
    </a>
  ) : (
    <span className={styles.itemName}>{item.n}</span>
  )

  return (
    <div className={styles.item}>
      {nameContent}
      {item.p != null ? (
        <span className={styles.itemPrice}>{formatPrice(item.p, item.c)}</span>
      ) : (
        <span className={styles.itemNoPrice}>—</span>
      )}
    </div>
  )
}

export default function SharedWishlistPage() {
  const [searchParams] = useSearchParams()
  const data = searchParams.get('data') ?? ''
  const items = decodeWishlist(data)
  const isInvalid = items.length === 0

  return (
    <div className={styles.page}>
      <div className={styles.container}>
        <header className={styles.header}>
          <span className={styles.badge}>SCOUTER</span>
          <h1 className={styles.title}>Liste partagée</h1>
          {!isInvalid && (
            <p className={styles.subtitle}>
              {items.length} article{items.length > 1 ? 's' : ''}
            </p>
          )}
        </header>

        {isInvalid ? (
          <div className={styles.error}>
            <p className={styles.errorTitle}>Lien invalide ou expiré</p>
          </div>
        ) : (
          <ul className={styles.list} role="list">
            {items.map((item, idx) => (
              <li key={idx}>
                <WishlistItemRow item={item} />
              </li>
            ))}
          </ul>
        )}

        <div className={styles.cta}>
          <a href="/wishlist" className={styles.ctaLink}>
            Créer ma liste Scouter
          </a>
        </div>
      </div>
    </div>
  )
}
