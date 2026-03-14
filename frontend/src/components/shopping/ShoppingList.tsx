import { MerchantGroup } from './MerchantGroup'
import { LoadingPulse } from '../scouter'
import type { ShoppingItem, ItemStatus } from '../../types'

interface ShoppingListProps {
  items: ShoppingItem[]
  isLoading?: boolean
  currency?: string
  onStatusChange?: (itemId: string, status: ItemStatus) => void
  onPriceClick?: (item: ShoppingItem) => void
}

export function ShoppingList({ items, isLoading, currency, onStatusChange, onPriceClick }: ShoppingListProps) {
  if (isLoading) return <LoadingPulse label="Loading items..." />
  if (items.length === 0) {
    return (
      <div style={{ padding: '3rem', textAlign: 'center', color: 'var(--text-dim)', fontSize: '0.85rem' }}>
        No items yet. Run Price Intel to populate this list.
      </div>
    )
  }

  const byMerchant = items.reduce<Record<string, ShoppingItem[]>>((acc, item) => {
    const key = item.merchant || 'Unknown'
    if (!acc[key]) acc[key] = []
    acc[key].push(item)
    return acc
  }, {})

  return (
    <div>
      {Object.entries(byMerchant).map(([merchant, merchantItems]) => (
        <MerchantGroup
          key={merchant}
          merchant={merchant}
          items={merchantItems}
          currency={currency}
          onStatusChange={onStatusChange}
          onPriceClick={onPriceClick}
        />
      ))}
    </div>
  )
}
