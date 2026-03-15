// ── Types ─────────────────────────────────────────────────────────────────────

export interface Envelope {
  id: string
  name: string
  emoji: string
  budgetEur: number
  spentEur: number
  color: string
  createdAt: string
}

export interface Transaction {
  id: string
  envelopeId: string
  label: string
  amountEur: number  // negative = expense, positive = refill
  date: string       // ISO
}

// ── Storage keys ──────────────────────────────────────────────────────────────

const ENVELOPES_KEY = 'scouter_envelopes'
const TRANSACTIONS_KEY = 'scouter_transactions'

// ── localStorage helpers ──────────────────────────────────────────────────────

export function loadEnvelopes(): Envelope[] {
  try {
    const raw = localStorage.getItem(ENVELOPES_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed as Envelope[]
  } catch {
    return []
  }
}

export function saveEnvelopes(envelopes: Envelope[]): void {
  localStorage.setItem(ENVELOPES_KEY, JSON.stringify(envelopes))
}

export function loadTransactions(): Transaction[] {
  try {
    const raw = localStorage.getItem(TRANSACTIONS_KEY)
    if (!raw) return []
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed as Transaction[]
  } catch {
    return []
  }
}

export function saveTransactions(transactions: Transaction[]): void {
  localStorage.setItem(TRANSACTIONS_KEY, JSON.stringify(transactions))
}

// ── Derived helpers ───────────────────────────────────────────────────────────

/** Sum of all expense transactions for an envelope (positive value = money spent) */
export function computeSpent(envelopeId: string, transactions: Transaction[]): number {
  return transactions
    .filter((t) => t.envelopeId === envelopeId)
    .reduce((sum, t) => sum - Math.min(t.amountEur, 0), 0)
}

// ── Preset colors ─────────────────────────────────────────────────────────────

export const PRESET_COLORS = [
  '#00e5ff',  // cyan
  '#00d68f',  // green
  '#ffd93d',  // gold
  '#f7974f',  // orange
  '#ff3d71',  // coral
  '#a855f7',  // purple
  '#38bdf8',  // sky
  '#fb7185',  // rose
]

export const DEFAULT_EMOJIS = [
  '🛒', '🍔', '🏠', '🚗', '💊', '🎬', '✈️', '📚', '👕', '💡', '🎮', '🐾', '🏋️', '🎁', '🔧', '💰',
]
