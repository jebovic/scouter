import { LoadingPulse } from '../scouter'
import { PriceHistoryChart } from './PriceHistoryChart'
import { usePriceSnapshots } from '../../hooks'
import type { ShoppingItem } from '../../types'
import styles from './PriceHistoryModal.module.css'

interface PriceHistoryModalProps {
  item: ShoppingItem
  currency: string
  onClose: () => void
}

export function PriceHistoryModal({ item, currency, onClose }: PriceHistoryModalProps) {
  const { snapshots, isLoading } = usePriceSnapshots(item.missionId, item.id)

  return (
    <div
      className={styles.overlay}
      onClick={(e) => e.target === e.currentTarget && onClose()}
    >
      <div
        className={styles.modal}
        role="dialog"
        aria-modal="true"
        aria-labelledby="price-history-modal-title"
      >
        <div className={styles.header}>
          <div>
            <h3 id="price-history-modal-title" className={styles.title}>PRICE HISTORY</h3>
            <p className={styles.subtitle}>{item.name} @ {item.merchant}</p>
          </div>
          <button className={styles.closeBtn} onClick={onClose} aria-label="Close price history">✕</button>
        </div>
        {isLoading ? (
          <LoadingPulse label="Loading history..." />
        ) : snapshots.length === 0 ? (
          <p className={styles.noHistory}>No price history yet</p>
        ) : (
          <PriceHistoryChart snapshots={snapshots} currency={currency} targetPrice={item.targetPrice} />
        )}
      </div>
    </div>
  )
}
