import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { MerchantGroup } from './MerchantGroup'
import { SmartSortToggle } from './SmartSortToggle'
import { LoadingPulse } from '../scouter'
import { useSortedItems } from '../../hooks/useSortedItems'
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
  const { t } = useTranslation()
  const [smartSort, setSmartSort] = useState(false)
  const { rankMap, isLoading: sortLoading } = useSortedItems(missionId, smartSort)

  // In smart-sort mode reorder items by rank; otherwise preserve original order.
  const displayItems = useMemo(() => {
    if (!smartSort || rankMap.size === 0) return items
    return [...items].sort((a, b) => {
      const ra = rankMap.get(a.id)?.rank ?? 9999
      const rb = rankMap.get(b.id)?.rank ?? 9999
      return ra - rb
    })
  }, [items, smartSort, rankMap])

  if (isLoading) return <LoadingPulse label={t('common.loading')} />
  if (items.length === 0) {
    return (
      <div className={styles.empty}>
        {t('shopping.noItems')}
      </div>
    )
  }

  const byMerchant = smartSort
    // In smart-sort mode flatten all items under a single virtual group to preserve rank order.
    ? { '': displayItems }
    : displayItems.reduce<Record<string, ShoppingItem[]>>((acc, item) => {
        const key = item.merchant || 'Unknown'
        return { ...acc, [key]: [...(acc[key] ?? []), item] }
      }, {})

  return (
    <div className={styles.root}>
      <div className={styles.toolbar}>
        <SmartSortToggle
          enabled={smartSort}
          onToggle={() => setSmartSort((v) => !v)}
          isLoading={sortLoading}
        />
      </div>
      {Object.entries(byMerchant).map(([merchant, merchantItems]) => (
        <MerchantGroup
          key={merchant || '__smart__'}
          missionId={missionId}
          merchant={merchant || ''}
          items={merchantItems}
          currency={currency}
          onStatusChange={onStatusChange}
          onPriceClick={onPriceClick}
          onPin={onPin}
          rankMap={smartSort ? rankMap : undefined}
        />
      ))}
    </div>
  )
}
