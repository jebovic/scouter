import { useState, useCallback } from 'react'

const STORAGE_KEY = 'scouter_loyalty_points'

export interface LoyaltyBalance {
  programId: string
  points: number
}

interface LoyaltyPointsState {
  balances: Record<string, number>
  setPoints: (programId: string, points: number) => void
  clearPoints: (programId: string) => void
}

function loadFromStorage(): Record<string, number> {
  try {
    const data = localStorage.getItem(STORAGE_KEY)
    return data ? JSON.parse(data) : {}
  } catch {
    return {}
  }
}

function saveToStorage(balances: Record<string, number>): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(balances))
  } catch {
    // ignore private browsing or quota exceeded
  }
}

export function useLoyaltyPoints(): LoyaltyPointsState {
  const [balances, setBalances] = useState<Record<string, number>>(() => loadFromStorage())

  const setPoints = useCallback((programId: string, points: number) => {
    setBalances((prev) => {
      const updated = { ...prev }
      if (points <= 0) {
        delete updated[programId]
      } else {
        updated[programId] = points
      }
      saveToStorage(updated)
      return updated
    })
  }, [])

  const clearPoints = useCallback((programId: string) => {
    setBalances((prev) => {
      const updated = { ...prev }
      delete updated[programId]
      saveToStorage(updated)
      return updated
    })
  }, [])

  return { balances, setPoints, clearPoints }
}
