import { useState, useCallback } from 'react'
import {
  loadEnvelopes,
  loadTransactions,
  saveEnvelopes,
  saveTransactions,
  computeSpent,
  type Envelope,
  type Transaction,
} from '../utils/envelopes'

function generateId(): string {
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
}

export function useLocalEnvelopes() {
  const [envelopes, setEnvelopes] = useState<Envelope[]>(() => loadEnvelopes())
  const [transactions, setTransactions] = useState<Transaction[]>(() => loadTransactions())

  // Derived: envelopes with live spentEur computed from transactions
  const enrichedEnvelopes: Envelope[] = envelopes.map((env) => ({
    ...env,
    spentEur: computeSpent(env.id, transactions),
  }))

  const addEnvelope = useCallback(
    (fields: Omit<Envelope, 'id' | 'spentEur' | 'createdAt'>) => {
      const next: Envelope = {
        ...fields,
        id: generateId(),
        spentEur: 0,
        createdAt: new Date().toISOString(),
      }
      setEnvelopes((prev) => {
        const updated = [...prev, next]
        saveEnvelopes(updated)
        return updated
      })
    },
    [],
  )

  const deleteEnvelope = useCallback((id: string) => {
    setEnvelopes((prev) => {
      const updated = prev.filter((e) => e.id !== id)
      saveEnvelopes(updated)
      return updated
    })
    // also remove associated transactions
    setTransactions((prev) => {
      const updated = prev.filter((t) => t.envelopeId !== id)
      saveTransactions(updated)
      return updated
    })
  }, [])

  const addTransaction = useCallback(
    (fields: Omit<Transaction, 'id'>) => {
      const next: Transaction = {
        ...fields,
        id: generateId(),
      }
      setTransactions((prev) => {
        const updated = [...prev, next]
        saveTransactions(updated)
        return updated
      })
    },
    [],
  )

  const deleteTransaction = useCallback((id: string) => {
    setTransactions((prev) => {
      const updated = prev.filter((t) => t.id !== id)
      saveTransactions(updated)
      return updated
    })
  }, [])

  return {
    envelopes: enrichedEnvelopes,
    transactions,
    addEnvelope,
    deleteEnvelope,
    addTransaction,
    deleteTransaction,
  }
}
