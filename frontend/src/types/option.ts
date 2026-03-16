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

export interface TranslationBlob {
  name?: string
  notes?: string
  warnings?: string[]
  category?: string
  attributes?: Array<{ label: string; value: string }>
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
  translations?: Record<string, TranslationBlob>
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
  name?: string
  badge?: OptionBadge
  attributes?: Attribute[]
  priceRange?: PriceRange
  notes?: string
  warnings?: string[]
}
