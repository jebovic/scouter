export interface Constraint {
  key: string
  label: string
  value: unknown
  type: 'hard' | 'soft'
}

export interface TimelineEvent {
  date: string
  label: string
  urgent: boolean
}

export type MissionCategory = 'travel' | 'electronics' | 'computing' | 'renovation' | 'custom'
export type MissionPhase = 'researching' | 'comparing' | 'buying' | 'done'

export interface WeightProfile {
  price: number
  quality: number
  feature: number
}

export interface Mission {
  id: string
  slug: string
  name: string
  icon: string
  category: MissionCategory
  autotagCategory?: string | null
  budget: number
  currency: string
  locale: string
  phase: MissionPhase
  constraints: Constraint[]
  costCategories: string[]
  timeline: TimelineEvent[]
  weightProfile: WeightProfile
  lessons?: string | null
  envelopeId?: string | null
  createdAt: string
  updatedAt: string
}

export interface MissionCreateRequest {
  name: string
  icon: string
  category: MissionCategory
  budget: number
  currency: string
  locale: string
  constraints: Constraint[]
  costCategories: string[]
}

export interface MissionUpdateRequest {
  name?: string
  icon?: string
  budget?: number
  phase?: MissionPhase
  constraints?: Constraint[]
  costCategories?: string[]
  timeline?: TimelineEvent[]
  weightProfile?: WeightProfile
  lessons?: string
  envelopeId?: string | null
}
