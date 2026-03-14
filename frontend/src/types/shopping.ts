export type ItemStatus = 'buy' | 'flash-sale' | 'preorder' | 'defer' | 'watch' | 'crisis' | 'recommended' | 'rejected'

export type PriceTrend = 'dropping' | 'stable' | 'rising'

export interface DealScore {
  score: number
  pctBelowAvg: number
  pctBelowTarget: number
  historicalAvg: number
  trend: PriceTrend
  snapshotCount: number
}

export interface ShoppingItem {
  id: string
  missionId: string
  name: string
  merchant: string
  costCategory: string
  price: number
  originalEstimate?: number
  targetPrice?: number
  status: ItemStatus
  note?: string
  url?: string
  pinned: boolean
  createdAt: string
}

export interface PriceSnapshot {
  id: string
  itemId: string
  price: number
  recordedAt: string
  note?: string
}

export interface ShoppingItemCreateRequest {
  name: string
  merchant: string
  costCategory: string
  price: number
  originalEstimate?: number
  status: ItemStatus
  note?: string
  url?: string
}

export interface ShoppingItemUpdateRequest {
  price?: number
  status?: ItemStatus
  note?: string
  merchant?: string
}

export interface PriceSnapshotRequest {
  price: number
  note?: string
}
