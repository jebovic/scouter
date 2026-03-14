import type { Constraint, WeightProfile, MissionCategory } from './mission'

export interface BudgetRange {
  min: number
  max: number
}

export interface Template {
  slug: string
  name: string
  icon: string
  category: MissionCategory
  description: string
  constraints: Constraint[]
  costCategories: string[]
  weightProfile: WeightProfile
  suggestedBudget?: BudgetRange
}
