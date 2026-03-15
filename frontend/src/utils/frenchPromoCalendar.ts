export interface PromoEvent {
  id: string
  name: string
  startMonth: number // 1-12
  startDay: number
  endMonth: number
  endDay: number
  category: string // 'soldes' | 'fetes' | 'tech' | 'mode' | 'food' | 'general'
  discount: string // e.g. "30-70%", "jusqu'à 50%"
  description: string
  icon: string
  color: string // CSS custom property name e.g. 'var(--gold)'
}

export const FRENCH_PROMO_EVENTS: PromoEvent[] = [
  {
    id: 'soldes-hiver',
    name: 'Soldes d\'Hiver',
    startMonth: 1,
    startDay: 8,
    endMonth: 2,
    endDay: 4,
    category: 'soldes',
    discount: '30-70%',
    description: 'Soldes réglementées post-fêtes',
    icon: '❄️',
    color: 'var(--accent)',
  },
  {
    id: 'saint-valentin',
    name: 'Saint-Valentin',
    startMonth: 2,
    startDay: 12,
    endMonth: 2,
    endDay: 14,
    category: 'fetes',
    discount: 'jusqu\'à 30%',
    description: 'Promos cadeaux et bijoux',
    icon: '❤️',
    color: 'var(--coral)',
  },
  {
    id: 'french-days',
    name: 'French Days',
    startMonth: 4,
    startDay: 28,
    endMonth: 5,
    endDay: 4,
    category: 'general',
    discount: 'jusqu\'à 50%',
    description: 'Alternative française au Black Friday',
    icon: '🇫🇷',
    color: 'var(--accent)',
  },
  {
    id: 'fete-meres',
    name: 'Fête des Mères',
    startMonth: 5,
    startDay: 23,
    endMonth: 5,
    endDay: 25,
    category: 'fetes',
    discount: 'jusqu\'à 40%',
    description: 'Promos cadeaux fête des mères',
    icon: '🌸',
    color: 'var(--gold)',
  },
  {
    id: 'soldes-ete',
    name: 'Soldes d\'Été',
    startMonth: 6,
    startDay: 25,
    endMonth: 7,
    endDay: 22,
    category: 'soldes',
    discount: '30-70%',
    description: 'Soldes réglementées été',
    icon: '☀️',
    color: 'var(--gold)',
  },
  {
    id: 'prime-day',
    name: 'Prime Day',
    startMonth: 7,
    startDay: 15,
    endMonth: 7,
    endDay: 16,
    category: 'tech',
    discount: 'jusqu\'à 60%',
    description: 'Ventes flash Amazon',
    icon: '⚡',
    color: 'var(--gold)',
  },
  {
    id: 'rentree',
    name: 'Rentrée',
    startMonth: 8,
    startDay: 15,
    endMonth: 9,
    endDay: 15,
    category: 'general',
    discount: 'jusqu\'à 30%',
    description: 'Promos fournitures et tech rentrée',
    icon: '🎒',
    color: 'var(--accent)',
  },
  {
    id: 'black-friday',
    name: 'Black Friday',
    startMonth: 11,
    startDay: 27,
    endMonth: 11,
    endDay: 28,
    category: 'general',
    discount: 'jusqu\'à 70%',
    description: 'Plus grande vente de l\'année',
    icon: '🖤',
    color: 'var(--text-dim)',
  },
  {
    id: 'cyber-monday',
    name: 'Cyber Monday',
    startMonth: 11,
    startDay: 30,
    endMonth: 12,
    endDay: 1,
    category: 'tech',
    discount: 'jusqu\'à 60%',
    description: 'Promos tech et high-tech',
    icon: '💻',
    color: 'var(--purple)',
  },
  {
    id: 'noel',
    name: 'Fêtes de Noël',
    startMonth: 12,
    startDay: 15,
    endMonth: 12,
    endDay: 25,
    category: 'fetes',
    discount: 'jusqu\'à 40%',
    description: 'Promos cadeaux et décorations',
    icon: '🎄',
    color: 'var(--green)',
  },
]

/**
 * Check if an event is active on a given date
 */
export function isEventActive(event: PromoEvent, date: Date): boolean {
  const month = date.getMonth() + 1
  const day = date.getDate()

  if (event.endMonth >= event.startMonth) {
    // Event does not wrap around year (e.g., Jan-Feb, May-May)
    if (month < event.startMonth || month > event.endMonth) {
      return false
    }
    if (month === event.startMonth && day < event.startDay) {
      return false
    }
    if (month === event.endMonth && day > event.endDay) {
      return false
    }
    return true
  } else {
    // Event wraps around year (e.g., Nov-Dec)
    if (month >= event.startMonth || month <= event.endMonth) {
      if (month === event.startMonth && day < event.startDay) {
        return false
      }
      if (month === event.endMonth && day > event.endDay) {
        return false
      }
      return true
    }
    return false
  }
}

/**
 * Get all promo events active on a given date
 */
export function getEventsForDate(date: Date): PromoEvent[] {
  return FRENCH_PROMO_EVENTS.filter((event) => isEventActive(event, date))
}

/**
 * Get upcoming promo events from a given date (next 60 days, including today if active)
 * Only includes events that are active or start within 60 days
 */
export function getUpcomingEvents(from: Date): PromoEvent[] {
  const upcomingSet = new Map<string, PromoEvent>()
  const year = from.getFullYear()

  // Check next 60 days
  for (let daysAhead = 0; daysAhead <= 60; daysAhead++) {
    const checkDate = new Date(from)
    checkDate.setDate(checkDate.getDate() + daysAhead)

    const eventsOnDate = getEventsForDate(checkDate)
    eventsOnDate.forEach((event) => {
      if (!upcomingSet.has(event.id)) {
        upcomingSet.set(event.id, event)
      }
    })
  }

  // Filter to only include events that haven't ended and sort by start date
  const currentDate = new Date(from)
  currentDate.setHours(0, 0, 0, 0)

  return Array.from(upcomingSet.values())
    .filter((event) => {
      // Include only if event ends on or after the from date
      const eventEndDate = new Date(year, event.endMonth - 1, event.endDay)
      // Handle year-wrapping events
      if (event.endMonth < event.startMonth) {
        // Event wraps to next year, check if we're past the end date
        if (currentDate.getMonth() + 1 < event.endMonth) {
          // We're not in the end month yet, so event is valid for next year
          return true
        }
        if (
          currentDate.getMonth() + 1 === event.endMonth &&
          currentDate.getDate() <= event.endDay
        ) {
          return true
        }
        // Check if event is still happening in the current year after wrapping
        if (currentDate.getMonth() + 1 >= event.startMonth) {
          return true
        }
        return false
      }
      // Non-wrapping events: check if end date hasn't passed
      return eventEndDate.getTime() >= currentDate.getTime()
    })
    .sort((a, b) => {
      const aDate = new Date(year, a.startMonth - 1, a.startDay)
      const bDate = new Date(year, b.startMonth - 1, b.startDay)

      // Handle year wrapping: if start month is earlier than current month, push to next year
      const currentMonth = currentDate.getMonth() + 1
      if (a.startMonth < currentMonth && a.startMonth < a.endMonth) {
        // Non-wrapping event in the past, push to next year
        aDate.setFullYear(aDate.getFullYear() + 1)
      }
      if (b.startMonth < currentMonth && b.startMonth < b.endMonth) {
        bDate.setFullYear(bDate.getFullYear() + 1)
      }

      return aDate.getTime() - bDate.getTime()
    })
}
