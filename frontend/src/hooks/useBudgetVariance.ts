import { useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import type { EnvelopeDTO } from '../api/envelope'

export interface VarianceCell {
  month: string // "YYYY-MM"
  monthLabel: string // "Jan 2025" (locale-aware)
  category: string // envelope name
  budgeted: number // monthly envelope amount
  spent: number // total purchases in that month for that category
  variance: number // spent - budgeted
  variancePct: number // (spent - budgeted) / budgeted * 100 (Infinity if no budget)
}

export interface BudgetVarianceData {
  months: string[] // last N months, oldest first
  categories: string[]
  cells: VarianceCell[]
}

/**
 * Generate month labels for the last N months in YYYY-MM format, oldest first.
 * Returns both the ISO format and the formatted label.
 */
function generateMonths(monthsBack: number, locale: string): { iso: string; label: string }[] {
  const result = []
  const now = new Date()

  // Generate from monthsBack months ago to now
  for (let i = monthsBack - 1; i >= 0; i--) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const iso = date.toISOString().slice(0, 7) // "YYYY-MM"
    const label = date.toLocaleDateString(locale, { month: 'short', year: 'numeric' })
    result.push({ iso, label })
  }

  return result
}

/**
 * Hook to derive monthly budget variance from envelopes.
 * Currently, spent is initialized to 0 (can be enhanced with actual transaction data later).
 * Returns the structure needed for the BudgetHeatmap component.
 */
export function useBudgetVariance(
  envelopes: EnvelopeDTO[] = [],
  monthsBack: number = 3,
): BudgetVarianceData {
  const { i18n } = useTranslation()
  const locale = i18n.language === 'fr' ? 'fr-FR' : 'en-US'

  return useMemo(() => {
    const months = generateMonths(monthsBack, locale)
    const monthIsos = months.map((m) => m.iso)
    const monthLabels = months.map((m) => m.label)

    const categories = envelopes.map((e) => e.name)

    // Create cells: one per month per category
    const cells: VarianceCell[] = []
    for (const envelope of envelopes) {
      for (let i = 0; i < monthIsos.length; i++) {
        const monthIso = monthIsos[i]
        const monthLabel = monthLabels[i]
        const budgeted = envelope.monthlyAmount
        const spent = 0 // Placeholder: will be enhanced with actual transaction data

        const variance = spent - budgeted
        const variancePct = budgeted > 0 ? (variance / budgeted) * 100 : Infinity

        cells.push({
          month: monthIso,
          monthLabel,
          category: envelope.name,
          budgeted,
          spent,
          variance,
          variancePct,
        })
      }
    }

    return {
      months: monthIsos,
      categories,
      cells,
    }
  }, [envelopes, monthsBack, locale])
}
