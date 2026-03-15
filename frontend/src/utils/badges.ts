export type BadgeId =
  | 'first_mission'
  | 'mission_completed'
  | 'five_missions'
  | 'saver_bronze'
  | 'saver_silver'
  | 'saver_gold'
  | 'researcher'
  | 'negotiator'
  | 'deal_hunter'
  | 'planner'

export interface Badge {
  id: BadgeId
  name: string
  description: string
  icon: string
  unlockedAt?: Date
}

export const ALL_BADGES: Record<BadgeId, Omit<Badge, 'unlockedAt'>> = {
  first_mission: {
    id: 'first_mission',
    name: 'Premier Éclaireur',
    description: 'Créer votre première mission',
    icon: '🎯',
  },
  mission_completed: {
    id: 'mission_completed',
    name: 'Mission Accomplie',
    description: 'Compléter une mission achat',
    icon: '✅',
  },
  five_missions: {
    id: 'five_missions',
    name: 'Commandant',
    description: 'Mener 5 missions',
    icon: '🏅',
  },
  saver_bronze: {
    id: 'saver_bronze',
    name: 'Épargnant Bronze',
    description: 'Économiser 100 €',
    icon: '🥉',
  },
  saver_silver: {
    id: 'saver_silver',
    name: 'Épargnant Argent',
    description: 'Économiser 500 €',
    icon: '🥈',
  },
  saver_gold: {
    id: 'saver_gold',
    name: 'Épargnant Or',
    description: 'Économiser 1 000 €',
    icon: '🥇',
  },
  researcher: {
    id: 'researcher',
    name: 'Analyste',
    description: 'Lancer 10 recherches IA',
    icon: '🔬',
  },
  negotiator: {
    id: 'negotiator',
    name: 'Négociateur',
    description: 'Utiliser le coach négo',
    icon: '🤝',
  },
  deal_hunter: {
    id: 'deal_hunter',
    name: 'Chasseur de Promo',
    description: 'Trouver une vente flash',
    icon: '⚡',
  },
  planner: {
    id: 'planner',
    name: 'Stratège',
    description: '5 options sur une mission',
    icon: '📋',
  },
}

/**
 * Determine which badges should be unlocked based on user metrics.
 * Returns array of BadgeIds that have been earned.
 */
export function unlockIfEarned(
  missionCount: number,
  savedTotal: number,
  researchCount: number
): BadgeId[] {
  const earned: BadgeId[] = []

  if (missionCount >= 1) {
    earned.push('first_mission')
    earned.push('mission_completed')
  }

  if (missionCount >= 5) {
    earned.push('five_missions')
  }

  if (savedTotal >= 100) {
    earned.push('saver_bronze')
  }

  if (savedTotal >= 500) {
    earned.push('saver_silver')
  }

  if (savedTotal >= 1000) {
    earned.push('saver_gold')
  }

  if (researchCount >= 10) {
    earned.push('researcher')
  }

  // Remove duplicates (immutable)
  return [...new Set(earned)]
}
