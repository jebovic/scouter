export type ItemStatus = 'buy' | 'flash-sale' | 'preorder' | 'defer' | 'watch' | 'crisis'

export interface ShoppingItem {
  id: string
  missionId: string
  name: string
  merchant: string
  costCategory: string
  price: number
  originalEstimate?: number
  status: ItemStatus
  note?: string
  url?: string
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
