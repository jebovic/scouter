import { useToggleWishListAlert } from '../../hooks/useWishList'
import type { WishListItem as WishListItemType } from '../../api/wishlist'
import styles from './WishListItem.module.css'

interface WishListItemProps {
  item: WishListItemType
  onDelete: (id: string) => void
}

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

export function WishListItem({ item, onDelete }: WishListItemProps) {
  const { id, name, url, targetPrice, currency, notes, alertEnabled, lastCheckedPrice } = item
  const toggleAlertMutation = useToggleWishListAlert(id)

  function handleToggleAlert() {
    toggleAlertMutation.mutate(!alertEnabled)
  }

  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.titleRow}>
          <span className={styles.name}>{name}</span>
          {url && (
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className={styles.link}
              aria-label={`Open ${name} product page`}
            >
              ↗
            </a>
          )}
        </div>

        <div className={styles.actions}>
          <button
            className={`${styles.alertBtn} ${alertEnabled ? styles.alertActive : styles.alertInactive}`}
            onClick={handleToggleAlert}
            disabled={toggleAlertMutation.isPending}
            aria-label={alertEnabled ? `Disable price alert for ${name}` : `Enable price alert for ${name}`}
            title={alertEnabled ? 'Disable price alert' : 'Enable price alert'}
          >
            {alertEnabled ? '🔔' : '🔕'}
          </button>

          <button
            className={styles.deleteBtn}
            onClick={() => onDelete(id)}
            aria-label={`Delete ${name}`}
          >
            ×
          </button>
        </div>
      </div>

      {targetPrice !== null && targetPrice !== undefined && (
        <div className={styles.price}>
          {formatPrice(targetPrice, currency)}
        </div>
      )}

      {lastCheckedPrice !== null && lastCheckedPrice !== undefined && (
        <p className={styles.lastChecked}>
          Last checked: {formatPrice(lastCheckedPrice, currency)}
        </p>
      )}

      {notes && (
        <p className={styles.notes}>{notes}</p>
      )}
    </div>
  )
}
