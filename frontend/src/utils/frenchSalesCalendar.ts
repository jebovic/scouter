export type SaleEventType = 'soldes' | 'promo' | 'flash' | 'tech' | 'seasonal'

export interface SaleEvent {
  id: string
  name: string
  nameEn: string
  startDate: string // ISO date YYYY-MM-DD
  endDate: string   // ISO date YYYY-MM-DD
  discountRange: string // e.g. "20-70%"
  bestCategories: string[] // categories where this event is most relevant
  type: SaleEventType
}

export const FRENCH_SALE_EVENTS_2026: SaleEvent[] = [
  {
    id: 'soldes-hiver-2026',
    name: "Soldes d'Hiver 2026",
    nameEn: 'Winter Sales',
    startDate: '2026-01-08',
    endDate: '2026-02-04',
    discountRange: '30-70%',
    bestCategories: ['clothing', 'fashion', 'shoes', 'home', 'electronics'],
    type: 'soldes',
  },
  {
    id: 'saint-valentin-2026',
    name: 'Saint-Valentin',
    nameEn: "Valentine's Day",
    startDate: '2026-02-07',
    endDate: '2026-02-14',
    discountRange: '10-30%',
    bestCategories: ['jewelry', 'perfume', 'gifts', 'luxury'],
    type: 'promo',
  },
  {
    id: 'french-days-2026',
    name: 'French Days',
    nameEn: 'French Days',
    startDate: '2026-04-29',
    endDate: '2026-05-04',
    discountRange: '20-50%',
    bestCategories: ['electronics', 'tech', 'appliances', 'sports'],
    type: 'tech',
  },
  {
    id: 'fete-meres-2026',
    name: 'Fête des Mères',
    nameEn: "Mother's Day",
    startDate: '2026-05-24',
    endDate: '2026-05-31',
    discountRange: '10-25%',
    bestCategories: ['jewelry', 'perfume', 'gifts', 'flowers', 'books'],
    type: 'seasonal',
  },
  {
    id: 'soldes-ete-2026',
    name: "Soldes d'Été 2026",
    nameEn: 'Summer Sales',
    startDate: '2026-06-25',
    endDate: '2026-07-22',
    discountRange: '30-70%',
    bestCategories: ['clothing', 'fashion', 'outdoor', 'sports', 'travel'],
    type: 'soldes',
  },
  {
    id: 'prime-day-2026',
    name: 'Amazon Prime Day (estimé)',
    nameEn: 'Prime Day (est.)',
    startDate: '2026-07-14',
    endDate: '2026-07-15',
    discountRange: '20-60%',
    bestCategories: ['electronics', 'tech', 'home', 'appliances'],
    type: 'tech',
  },
  {
    id: 'back-to-school-2026',
    name: 'Rentrée Scolaire',
    nameEn: 'Back to School',
    startDate: '2026-08-15',
    endDate: '2026-09-15',
    discountRange: '10-30%',
    bestCategories: ['electronics', 'tech', 'books', 'stationery', 'furniture'],
    type: 'seasonal',
  },
  {
    id: 'black-friday-2026',
    name: 'Black Friday',
    nameEn: 'Black Friday',
    startDate: '2026-11-27',
    endDate: '2026-11-27',
    discountRange: '20-60%',
    bestCategories: ['electronics', 'tech', 'appliances', 'clothing', 'home'],
    type: 'flash',
  },
  {
    id: 'cyber-monday-2026',
    name: 'Cyber Monday',
    nameEn: 'Cyber Monday',
    startDate: '2026-11-30',
    endDate: '2026-11-30',
    discountRange: '20-50%',
    bestCategories: ['electronics', 'tech', 'software', 'gaming'],
    type: 'tech',
  },
  {
    id: 'noel-2026',
    name: 'Promotions Noël',
    nameEn: 'Christmas Sales',
    startDate: '2026-12-01',
    endDate: '2026-12-24',
    discountRange: '10-40%',
    bestCategories: ['toys', 'gifts', 'electronics', 'books', 'games'],
    type: 'seasonal',
  },
]

/** Returns upcoming events in the next 90 days from today */
export function getUpcomingEvents(today = new Date()): SaleEvent[] {
  const cutoff = new Date(today)
  cutoff.setDate(cutoff.getDate() + 90)
  return FRENCH_SALE_EVENTS_2026.filter((e) => {
    const end = new Date(e.endDate)
    const start = new Date(e.startDate)
    return end >= today && start <= cutoff
  }).sort((a, b) => a.startDate.localeCompare(b.startDate))
}

/** Returns events most relevant for a given category */
export function getEventsForCategory(category: string): SaleEvent[] {
  const cat = category.toLowerCase()
  return FRENCH_SALE_EVENTS_2026.filter((e) =>
    e.bestCategories.some(
      (c) => c.toLowerCase().includes(cat) || cat.includes(c.toLowerCase()),
    ),
  ).sort((a, b) => a.startDate.localeCompare(b.startDate))
}

/** Returns the next best sale event for a category */
export function getNextBestEvent(category: string, today = new Date()): SaleEvent | null {
  const events = getEventsForCategory(category)
  return events.find((e) => new Date(e.endDate) >= today) ?? null
}

/** Returns days until an event start (negative if already started) */
export function daysUntil(dateStr: string, today = new Date()): number {
  const target = new Date(dateStr)
  const diff = target.getTime() - today.getTime()
  return Math.ceil(diff / (1000 * 60 * 60 * 24))
}
