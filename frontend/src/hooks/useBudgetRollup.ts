import { useMemo } from 'react'
import { useMissions } from './useMission'

export interface MissionBudgetEntry {
  missionId: string
  missionName: string
  missionSlug: string
  category: string
  budget: number
  phase: string
  currency: string
}

export interface BudgetRollup {
  totalBudget: number
  byCategory: Record<string, number>
  byPhase: Record<string, number>
  activeMissions: number
  purchasedTotal: number
  entries: MissionBudgetEntry[]
}

const EMPTY_ROLLUP: BudgetRollup = {
  totalBudget: 0,
  byCategory: {},
  byPhase: {},
  activeMissions: 0,
  purchasedTotal: 0,
  entries: [],
}

export function useBudgetRollup(): { rollup: BudgetRollup; isLoading: boolean } {
  const { missions, isLoading } = useMissions()

  const rollup = useMemo<BudgetRollup>(() => {
    if (!missions || missions.length === 0) return EMPTY_ROLLUP

    // Filter to missions with budget > 0
    const withBudget = missions.filter((m) => m.budget > 0)

    if (withBudget.length === 0) return EMPTY_ROLLUP

    const entries: MissionBudgetEntry[] = withBudget.map((m) => ({
      missionId: m.id,
      missionName: m.name,
      missionSlug: m.slug,
      category: m.category,
      budget: m.budget,
      phase: m.phase,
      currency: m.currency ?? 'EUR',
    }))

    // totalBudget excludes 'done' missions
    const totalBudget = entries
      .filter((e) => e.phase !== 'done')
      .reduce((sum, e) => sum + e.budget, 0)

    // purchasedTotal = sum of 'done' missions
    const purchasedTotal = entries
      .filter((e) => e.phase === 'done')
      .reduce((sum, e) => sum + e.budget, 0)

    // activeMissions = count of non-done missions with budget > 0
    const activeMissions = entries.filter((e) => e.phase !== 'done').length

    // byCategory — all missions with budget > 0
    const byCategory: Record<string, number> = {}
    for (const e of entries) {
      byCategory[e.category] = (byCategory[e.category] ?? 0) + e.budget
    }

    // byPhase — all missions with budget > 0
    const byPhase: Record<string, number> = {}
    for (const e of entries) {
      byPhase[e.phase] = (byPhase[e.phase] ?? 0) + e.budget
    }

    return {
      totalBudget,
      byCategory,
      byPhase,
      activeMissions,
      purchasedTotal,
      entries,
    }
  }, [missions])

  return { rollup, isLoading }
}
