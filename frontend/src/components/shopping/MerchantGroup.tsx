import { ShoppingItemRow } from './ShoppingItemRow'
import type { ShoppingItem, ItemStatus } from '../../types'
import styles from './MerchantGroup.module.css'

interface MerchantGroupProps {
  missionId: string
  merchant: string
  items: ShoppingItem[]
  currency?: string
  onStatusChange?: (itemId: string, status: ItemStatus) => void
  onPriceClick?: (item: ShoppingItem) => void
  onPin?: (itemId: string) => void
}

export function MerchantGroup({ missionId, merchant, items, currency = 'USD', onStatusChange, onPriceClick, onPin }: MerchantGroupProps) {
  const total = items.reduce((sum, i) => sum + i.price, 0)
  const fmt = (n: number) =>
    new Intl.NumberFormat(undefined, { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

  return (
    <div className={styles.root}>
      <div className={styles.header}>
        <span className={styles.merchantName}>{merchant}</span>
        <span className={styles.total}>{fmt(total)}</span>
      </div>
      {items.map((item) => (
        <ShoppingItemRow
          key={item.id}
          item={item}
          missionId={missionId}
          currency={currency}
          onStatusChange={onStatusChange ? (s) => onStatusChange(item.id, s) : undefined}
          onPriceClick={onPriceClick ? () => onPriceClick(item) : undefined}
          onPin={onPin}
        />
      ))}
    </div>
  )
}
