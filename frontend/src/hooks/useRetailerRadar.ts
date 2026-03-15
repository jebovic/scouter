import { useMemo } from 'react'
import { useShopping } from './useShopping'

export interface RetailerStat {
  merchant: string
  bestPrice: number
  currency: string
  itemName: string
  itemId: string
  inStock: boolean
}

export function useRetailerRadar(missionId: string) {
  const { items } = useShopping(missionId)

  const stats: RetailerStat[] = useMemo(() => {
    const map = new Map<string, RetailerStat>()

    for (const item of items) {
      // Skip items without merchant or invalid price
      if (!item.merchant || item.price <= 0) continue

      const existing = map.get(item.merchant)

      // Keep lowest price per merchant
      if (!existing || item.price < existing.bestPrice) {
        map.set(item.merchant, {
          merchant: item.merchant,
          bestPrice: item.price,
          currency: 'EUR',
          itemName: item.name,
          itemId: item.id,
          inStock: true,
        })
      }
    }

    // Sort by price ascending
    return Array.from(map.values()).sort((a, b) => a.bestPrice - b.bestPrice)
  }, [items])

  return { stats }
}
