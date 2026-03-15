export interface SeasonalPeriod {
  month: number // 1-12
  label: string // e.g. "Jan", "Fév"
  priceLevel: 'low' | 'medium' | 'high' // relative to annual average
  events: string[] // e.g. ["Soldes d'hiver", "French Days"]
}

export interface CategorySeason {
  category: string
  icon: string
  bestMonths: number[] // months to buy
  worstMonths: number[] // months to avoid
  tip: string // short French tip
}

export const FRENCH_MONTHS = [
  'Jan', 'Fév', 'Mar', 'Avr', 'Mai', 'Jun',
  'Jul', 'Aoû', 'Sep', 'Oct', 'Nov', 'Déc',
]

export const CATEGORY_SEASONS: CategorySeason[] = [
  {
    category: 'Électronique',
    icon: '📱',
    bestMonths: [7, 11],
    worstMonths: [3, 12],
    tip: 'Achetez en juillet (Soldes d\'été) ou novembre (Black Friday) pour les meilleures réductions.',
  },
  {
    category: 'TV & Audio',
    icon: '📺',
    bestMonths: [6, 7, 11],
    worstMonths: [1, 12],
    tip: 'Les Soldes d\'été et Black Friday sont les meilleurs moments pour les téléviseurs.',
  },
  {
    category: 'Electroménager',
    icon: '🏠',
    bestMonths: [1, 6, 7],
    worstMonths: [11, 12],
    tip: 'Achetez en janvier (Soldes d\'hiver) ou en été pour l\'électroménager.',
  },
  {
    category: 'Mode',
    icon: '👗',
    bestMonths: [1, 7],
    worstMonths: [3, 9],
    tip: 'Les Soldes d\'hiver et d\'été offrent les meilleures réductions sur la mode.',
  },
  {
    category: 'Jouets',
    icon: '🎮',
    bestMonths: [1, 7, 10],
    worstMonths: [11, 12],
    tip: 'Évitez novembre-décembre. Janvier et octobre offrent de bonnes réductions.',
  },
  {
    category: 'Sport',
    icon: '⚽',
    bestMonths: [1, 6, 7, 9],
    worstMonths: [4, 5],
    tip: 'Achetez en janvier ou lors des Soldes d\'été pour l\'équipement sportif.',
  },
  {
    category: 'Jardinage',
    icon: '🌱',
    bestMonths: [3, 4, 5],
    worstMonths: [7, 8],
    tip: 'Le printemps est la meilleure saison pour les équipements de jardinage.',
  },
  {
    category: 'Voyages',
    icon: '✈️',
    bestMonths: [1, 5, 9],
    worstMonths: [7, 8, 12],
    tip: 'Évitez l\'été et décembre. Janvier et septembre offrent les meilleurs tarifs.',
  },
  {
    category: 'Auto',
    icon: '🚗',
    bestMonths: [1, 8, 12],
    worstMonths: [4],
    tip: 'Les fins de mois et de trimestres offrent les meilleures négociations.',
  },
  {
    category: 'Livres',
    icon: '📚',
    bestMonths: [1, 7, 9],
    worstMonths: [12],
    tip: 'Achetez en janvier, septembre ou lors des Soldes d\'été pour les livres.',
  },
]

export function getCurrentMonthRecommendations(month: number): CategorySeason[] {
  return CATEGORY_SEASONS.filter((cat) => cat.bestMonths.includes(month))
}
