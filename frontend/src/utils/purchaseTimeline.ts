export interface TimelineItem {
  itemId: string
  itemName: string
  merchant: string
  currentPrice: number
  targetPrice?: number
  // Recommended purchase window
  windowStart: Date // e.g. "this week", "next month"
  windowEnd: Date
  priority: 'now' | 'soon' | 'wait' // now=within week, soon=within month, wait=2+ months
  reason: string // French explanation
}

export interface ItemInput {
  id: string
  name: string
  merchant: string
  price: number
  targetPrice?: number | null
  dealScore?: number
}

/**
 * Compute purchase timeline for shopping items based on price targets and deal scores.
 *
 * Logic:
 * - If price <= targetPrice (or no targetPrice): priority='now', windowStart=today, windowEnd=today+7d, reason="Prix cible atteint"
 * - If dealScore >= 80: priority='now', windowStart=today, windowEnd=today+3d, reason="Excellent deal score"
 * - If dealScore >= 60: priority='soon', windowStart=today+7d, windowEnd=today+30d, reason="Bon deal, à surveiller"
 * - Otherwise: priority='wait', windowStart=today+30d, windowEnd=today+90d, reason="Attendre une meilleure opportunité"
 */
export function computePurchaseTimeline(
  items: ItemInput[],
  referenceDate?: Date,
): TimelineItem[] {
  const now = referenceDate ?? new Date()

  return items.map((item) => {
    const isTargetMet = item.targetPrice !== undefined && item.targetPrice !== null && item.price <= item.targetPrice
    const dealScore = item.dealScore ?? 0
    const hasExcellentDeal = dealScore >= 80
    const hasGoodDeal = dealScore >= 60

    let priority: 'now' | 'soon' | 'wait'
    let reason: string
    let windowStart: Date
    let windowEnd: Date

    if (isTargetMet) {
      priority = 'now'
      reason = 'Prix cible atteint'
      windowStart = new Date(now)
      windowEnd = addDays(now, 7)
    } else if (hasExcellentDeal) {
      priority = 'now'
      reason = 'Excellent deal score'
      windowStart = new Date(now)
      windowEnd = addDays(now, 3)
    } else if (hasGoodDeal) {
      priority = 'soon'
      reason = 'Bon deal, à surveiller'
      windowStart = addDays(now, 7)
      windowEnd = addDays(now, 30)
    } else {
      priority = 'wait'
      reason = 'Attendre une meilleure opportunité'
      windowStart = addDays(now, 30)
      windowEnd = addDays(now, 90)
    }

    return {
      itemId: item.id,
      itemName: item.name,
      merchant: item.merchant,
      currentPrice: item.price,
      targetPrice: item.targetPrice ?? undefined,
      windowStart,
      windowEnd,
      priority,
      reason,
    }
  })
}

function addDays(date: Date, days: number): Date {
  const result = new Date(date)
  result.setDate(result.getDate() + days)
  return result
}
