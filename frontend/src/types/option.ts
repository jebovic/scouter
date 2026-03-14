export type AttributeType = 'text' | 'price' | 'score' | 'boolean'
export type OptionBadge = 'recommended' | 'alternative' | 'rejected' | 'watch'

export interface Attribute {
  key: string
  label: string
  value: unknown
  type: AttributeType
  max?: number
  pass?: boolean
}

export interface PriceRange {
  min: number
  max: number
  best: number
}

export interface Option {
  id: string
  missionId: string
  name: string
  category: string
  badge: OptionBadge
  attributes: Attribute[]
  priceRange?: PriceRange
  notes?: string
  warnings: string[]
  url?: string
  pinned: boolean
  rejected: boolean
  rejectReason?: string
  createdAt: string
}

export interface OptionCreateRequest {
  name: string
  category: string
  badge: OptionBadge
  attributes: Attribute[]
  priceRange?: PriceRange
  notes?: string
  warnings: string[]
  url?: string
}

export interface OptionUpdateRequest {
  badge?: OptionBadge
  attributes?: Attribute[]
  priceRange?: PriceRange
  notes?: string
  warnings?: string[]
}
