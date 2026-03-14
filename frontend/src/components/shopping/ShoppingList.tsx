import { MerchantGroup } from './MerchantGroup'
import { LoadingPulse } from '../scouter'
import type { ShoppingItem, ItemStatus } from '../../types'
import styles from './ShoppingList.module.css'

interface ShoppingListProps {
  missionId: string
  items: ShoppingItem[]
  isLoading?: boolean
  currency?: string
  onStatusChange?: (itemId: string, status: ItemStatus) => void
  onPriceClick?: (item: ShoppingItem) => void
  onPin?: (itemId: string) => void
}

export function ShoppingList({ missionId, items, isLoading, currency, onStatusChange, onPriceClick, onPin }: ShoppingListProps) {
  if (isLoading) return <LoadingPulse label="Loading items..." />
  if (items.length === 0) {
    return (
      <div className={styles.empty}>
        No items yet. Run Price Intel to populate this list.
      </div>
    )
  }

  const byMerchant = items.reduce<Record<string, ShoppingItem[]>>((acc, item) => {
    const key = item.merchant || 'Unknown'
    return { ...acc, [key]: [...(acc[key] ?? []), item] }
  }, {})

  return (
    <div className={styles.root}>
      {Object.entries(byMerchant).map(([merchant, merchantItems]) => (
        <MerchantGroup
          key={merchant}
          missionId={missionId}
          merchant={merchant}
          items={merchantItems}
          currency={currency}
          onStatusChange={onStatusChange}
          onPriceClick={onPriceClick}
          onPin={onPin}
        />
      ))}
    </div>
  )
}
