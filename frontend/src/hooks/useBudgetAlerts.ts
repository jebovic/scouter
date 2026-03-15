import { useEffect, useRef } from 'react'
import { detectBudgetAlerts, BudgetAlert } from '../utils/budgetAlerts'

export function useBudgetAlerts(
  spent: number,
  budget: number,
  itemCount: number,
  onAlert: (alert: BudgetAlert) => void,
) {
  const prevSpentRef = useRef<number | undefined>(undefined)
  const shownRef = useRef<Set<string>>(new Set())

  useEffect(() => {
    if (budget <= 0) return
    const alerts = detectBudgetAlerts(spent, budget, itemCount, prevSpentRef.current)
    for (const alert of alerts) {
      // Deduplicate by type per session
      const key = alert.type + '-' + Math.floor(spent / 100)
      if (!shownRef.current.has(key)) {
        shownRef.current.add(key)
        onAlert(alert)
      }
    }
    prevSpentRef.current = spent
  }, [spent, budget, itemCount, onAlert])
}
