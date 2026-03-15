export interface SpendingPattern {
  category: string
  totalSpent: number
  itemCount: number
  avgPrice: number
  pctOfBudget: number
}

export interface SpendingInsight {
  type: 'tip' | 'warning' | 'achievement'
  title: string
  description: string
  emoji: string
}

export interface SpendingReport {
  totalSpent: number
  totalBudget: number
  utilizationPct: number
  topCategory: string | null
  patterns: SpendingPattern[]
  insights: SpendingInsight[]
  savingsEstimate: number
}

import { formatCurrency } from './currency'

export function analyzeSpending(
  missions: Array<{ name: string; category: string; budget: number }>,
  allItems: Array<{ price: number; costCategory: string; targetPrice?: number | null }>,
  currency = 'EUR',
  locale = 'en-US',
): SpendingReport {
  const totalSpent = allItems.reduce((s, i) => s + i.price, 0)
  const totalBudget = missions.reduce((s, m) => s + m.budget, 0)
  const utilizationPct = totalBudget > 0 ? (totalSpent / totalBudget) * 100 : 0

  // Group by category
  const byCategory = new Map<string, { spent: number; count: number }>()
  for (const item of allItems) {
    const cat = item.costCategory || 'Autre'
    const existing = byCategory.get(cat) ?? { spent: 0, count: 0 }
    byCategory.set(cat, { spent: existing.spent + item.price, count: existing.count + 1 })
  }

  const patterns: SpendingPattern[] = Array.from(byCategory.entries())
    .map(([category, { spent, count }]) => ({
      category,
      totalSpent: spent,
      itemCount: count,
      avgPrice: count > 0 ? spent / count : 0,
      pctOfBudget: totalBudget > 0 ? (spent / totalBudget) * 100 : 0,
    }))
    .sort((a, b) => b.totalSpent - a.totalSpent)

  const topCategory = patterns[0]?.category ?? null

  // Savings estimate: sum of (price - targetPrice) where targetPrice < price
  const savingsEstimate = allItems.reduce((s, i) => {
    if (i.targetPrice != null && i.targetPrice < i.price) {
      return s + (i.price - i.targetPrice)
    }
    return s
  }, 0)

  const insights: SpendingInsight[] = []

  if (utilizationPct > 90) {
    insights.push({
      type: 'warning',
      title: 'Budget presque épuisé',
      description: `${utilizationPct.toFixed(0)}% de votre budget global est utilisé.`,
      emoji: '⚠️',
    })
  } else if (utilizationPct < 30 && totalBudget > 0) {
    insights.push({
      type: 'achievement',
      title: 'Excellente maîtrise',
      description: 'Moins de 30% de budget utilisé — vous êtes en bonne voie!',
      emoji: '🏆',
    })
  }

  if (savingsEstimate > 0) {
    insights.push({
      type: 'tip',
      title: 'Économies potentielles',
      description: `Attendez vos prix cibles pour économiser ${formatCurrency(savingsEstimate, currency, locale)}.`,
      emoji: '💡',
    })
  }

  if (topCategory) {
    insights.push({
      type: 'tip',
      title: `Catégorie dominante: ${topCategory}`,
      description: `${((patterns[0].totalSpent / (totalSpent || 1)) * 100).toFixed(0)}% de vos dépenses sont dans "${topCategory}".`,
      emoji: '📊',
    })
  }

  if (missions.length >= 3) {
    insights.push({
      type: 'achievement',
      title: 'Scout actif',
      description: `${missions.length} missions en cours — vous comparez intelligemment!`,
      emoji: '🎯',
    })
  }

  return { totalSpent, totalBudget, utilizationPct, topCategory, patterns, insights, savingsEstimate }
}
