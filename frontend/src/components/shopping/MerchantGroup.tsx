import { ShoppingItemRow } from './ShoppingItemRow'
import { formatCurrency } from '../../utils/format'
import type { ShoppingItem, ItemStatus } from '../../types'
import type { RankedItem } from '../../api/listsorter'
import styles from './MerchantGroup.module.css'

interface MerchantGroupProps {
  missionId: string
  merchant: string
  items: ShoppingItem[]
  currency?: string
  onStatusChange?: (itemId: string, status: ItemStatus) => void
  onPriceClick?: (item: ShoppingItem) => void
  onPin?: (itemId: string) => void
  /** When provided, ShoppingItemRow will show rank badges */
  rankMap?: Map<string, RankedItem>
}

export function MerchantGroup({ missionId, merchant, items, currency = 'USD', onStatusChange, onPriceClick, onPin, rankMap }: MerchantGroupProps) {
  const total = items.reduce((sum, i) => sum + i.price, 0)

  return (
    <div className={styles.root}>
      {merchant && (
        <div className={styles.header}>
          <span className={styles.merchantName}>{merchant}</span>
          <span className={styles.total}>{formatCurrency(total, currency)}</span>
        </div>
      )}
      {items.map((item) => (
        <ShoppingItemRow
          key={item.id}
          item={item}
          missionId={missionId}
          currency={currency}
          onStatusChange={onStatusChange ? (s) => onStatusChange(item.id, s) : undefined}
          onPriceClick={onPriceClick ? () => onPriceClick(item) : undefined}
          onPin={onPin}
          rank={rankMap?.get(item.id)}
        />
      ))}
    </div>
  )
}
