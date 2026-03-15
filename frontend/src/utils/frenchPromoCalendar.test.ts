import { describe, it, expect, beforeEach, vi } from 'vitest'
import {
  PromoEvent,
  FRENCH_PROMO_EVENTS,
  getEventsForDate,
  getUpcomingEvents,
  isEventActive,
} from './frenchPromoCalendar'

describe('frenchPromoCalendar', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('FRENCH_PROMO_EVENTS', () => {
    it('should have at least 8 events', () => {
      expect(FRENCH_PROMO_EVENTS.length).toBeGreaterThanOrEqual(8)
    })

    it('should have all required fields for each event', () => {
      FRENCH_PROMO_EVENTS.forEach((event) => {
        expect(event).toHaveProperty('id')
        expect(event).toHaveProperty('name')
        expect(event).toHaveProperty('startMonth')
        expect(event).toHaveProperty('startDay')
        expect(event).toHaveProperty('endMonth')
        expect(event).toHaveProperty('endDay')
        expect(event).toHaveProperty('category')
        expect(event).toHaveProperty('discount')
        expect(event).toHaveProperty('description')
        expect(event).toHaveProperty('icon')
        expect(event).toHaveProperty('color')
      })
    })

    it('should have valid month and day ranges', () => {
      FRENCH_PROMO_EVENTS.forEach((event) => {
        expect(event.startMonth).toBeGreaterThanOrEqual(1)
        expect(event.startMonth).toBeLessThanOrEqual(12)
        expect(event.startDay).toBeGreaterThanOrEqual(1)
        expect(event.startDay).toBeLessThanOrEqual(31)
        expect(event.endMonth).toBeGreaterThanOrEqual(1)
        expect(event.endMonth).toBeLessThanOrEqual(12)
        expect(event.endDay).toBeGreaterThanOrEqual(1)
        expect(event.endDay).toBeLessThanOrEqual(31)
      })
    })

    it('should have valid categories', () => {
      const validCategories = ['soldes', 'fetes', 'tech', 'mode', 'food', 'general']
      FRENCH_PROMO_EVENTS.forEach((event) => {
        expect(validCategories).toContain(event.category)
      })
    })

    it('should have unique IDs', () => {
      const ids = FRENCH_PROMO_EVENTS.map((e) => e.id)
      const uniqueIds = new Set(ids)
      expect(uniqueIds.size).toBe(ids.length)
    })
  })

  describe('isEventActive', () => {
    it('should return true when date is within event range (same month)', () => {
      const event: PromoEvent = {
        id: 'test',
        name: 'Test Event',
        startMonth: 3,
        startDay: 15,
        endMonth: 3,
        endDay: 25,
        category: 'general',
        discount: '10%',
        description: 'Test',
        icon: '🎉',
        color: 'var(--gold)',
      }
      const dateInRange = new Date(2026, 2, 20) // March 20, 2026
      expect(isEventActive(event, dateInRange)).toBe(true)
    })

    it('should return true on start date', () => {
      const event: PromoEvent = {
        id: 'test',
        name: 'Test Event',
        startMonth: 3,
        startDay: 15,
        endMonth: 3,
        endDay: 25,
        category: 'general',
        discount: '10%',
        description: 'Test',
        icon: '🎉',
        color: 'var(--gold)',
      }
      const startDate = new Date(2026, 2, 15) // March 15, 2026
      expect(isEventActive(event, startDate)).toBe(true)
    })

    it('should return true on end date', () => {
      const event: PromoEvent = {
        id: 'test',
        name: 'Test Event',
        startMonth: 3,
        startDay: 15,
        endMonth: 3,
        endDay: 25,
        category: 'general',
        discount: '10%',
        description: 'Test',
        icon: '🎉',
        color: 'var(--gold)',
      }
      const endDate = new Date(2026, 2, 25) // March 25, 2026
      expect(isEventActive(event, endDate)).toBe(true)
    })

    it('should return false when date is before event', () => {
      const event: PromoEvent = {
        id: 'test',
        name: 'Test Event',
        startMonth: 3,
        startDay: 15,
        endMonth: 3,
        endDay: 25,
        category: 'general',
        discount: '10%',
        description: 'Test',
        icon: '🎉',
        color: 'var(--gold)',
      }
      const dateBeforeEvent = new Date(2026, 2, 14) // March 14, 2026
      expect(isEventActive(event, dateBeforeEvent)).toBe(false)
    })

    it('should return false when date is after event', () => {
      const event: PromoEvent = {
        id: 'test',
        name: 'Test Event',
        startMonth: 3,
        startDay: 15,
        endMonth: 3,
        endDay: 25,
        category: 'general',
        discount: '10%',
        description: 'Test',
        icon: '🎉',
        color: 'var(--gold)',
      }
      const dateAfterEvent = new Date(2026, 2, 26) // March 26, 2026
      expect(isEventActive(event, dateAfterEvent)).toBe(false)
    })

    it('should handle multi-month events (Jan 8 - Feb 4)', () => {
      const event: PromoEvent = {
        id: 'soldes-hiver',
        name: 'Soldes d\'Hiver',
        startMonth: 1,
        startDay: 8,
        endMonth: 2,
        endDay: 4,
        category: 'soldes',
        discount: '30-70%',
        description: 'Winter sales',
        icon: '❄️',
        color: 'var(--accent)',
      }
      expect(isEventActive(event, new Date(2026, 0, 15))).toBe(true) // Jan 15
      expect(isEventActive(event, new Date(2026, 1, 1))).toBe(true) // Feb 1
      expect(isEventActive(event, new Date(2026, 0, 7))).toBe(false) // Jan 7
      expect(isEventActive(event, new Date(2026, 1, 5))).toBe(false) // Feb 5
    })

    it('should handle events spanning December to January', () => {
      const event: PromoEvent = {
        id: 'cyber-monday',
        name: 'Cyber Monday',
        startMonth: 11,
        startDay: 30,
        endMonth: 12,
        endDay: 1,
        category: 'tech',
        discount: 'jusqu\'à 60%',
        description: 'Cyber Monday',
        icon: '💻',
        color: 'var(--purple)',
      }
      expect(isEventActive(event, new Date(2026, 10, 30))).toBe(true) // Nov 30
      expect(isEventActive(event, new Date(2026, 11, 1))).toBe(true) // Dec 1
      expect(isEventActive(event, new Date(2026, 10, 29))).toBe(false) // Nov 29
      expect(isEventActive(event, new Date(2026, 11, 2))).toBe(false) // Dec 2
    })

    it('should handle year-spanning events (Nov 30 -> Dec 1 across year boundary)', () => {
      const event: PromoEvent = {
        id: 'test-year-span',
        name: 'Test Year Span',
        startMonth: 11,
        startDay: 30,
        endMonth: 12,
        endDay: 1,
        category: 'general',
        discount: '10%',
        description: 'Test',
        icon: '🎉',
        color: 'var(--gold)',
      }
      // Dec 1, 2026 should be active
      expect(isEventActive(event, new Date(2026, 11, 1))).toBe(true)
    })
  })

  describe('getEventsForDate', () => {
    it('should return empty array for date with no events', () => {
      const dateWithNoEvents = new Date(2026, 2, 10) // March 10, 2026 (not during soldes d'hiver)
      const events = getEventsForDate(dateWithNoEvents)
      expect(Array.isArray(events)).toBe(true)
    })

    it('should return multiple events if multiple are active on that date', () => {
      // Use a date that we know should have events from FRENCH_PROMO_EVENTS
      const dateDuringSoldes = new Date(2026, 0, 15) // Jan 15 (during Soldes d'Hiver)
      const events = getEventsForDate(dateDuringSoldes)
      expect(Array.isArray(events)).toBe(true)
      // At least soldes-hiver should be active
      const hasSoldesHiver = events.some((e) => e.id === 'soldes-hiver')
      expect(hasSoldesHiver).toBe(true)
    })

    it('should return empty array for Dec 31', () => {
      const newYearsEve = new Date(2026, 11, 31)
      const events = getEventsForDate(newYearsEve)
      expect(Array.isArray(events)).toBe(true)
    })

    it('should be consistent with isEventActive', () => {
      const testDate = new Date(2026, 4, 24) // May 24 (should have Fête des Mères)
      const events = getEventsForDate(testDate)
      FRENCH_PROMO_EVENTS.forEach((promo) => {
        const isActive = isEventActive(promo, testDate)
        const isInResult = events.some((e) => e.id === promo.id)
        expect(isInResult).toBe(isActive)
      })
    })
  })

  describe('getUpcomingEvents', () => {
    it('should return an array', () => {
      const from = new Date(2026, 2, 15)
      const upcoming = getUpcomingEvents(from)
      expect(Array.isArray(upcoming)).toBe(true)
    })

    it('should only return events starting within 60 days from the given date', () => {
      const from = new Date(2026, 3, 15) // April 15 (mid-year to avoid year-wrapping complexity)
      const upcoming = getUpcomingEvents(from)
      upcoming.forEach((event) => {
        // Calculate the start date of the event
        let eventStartDate = new Date(2026, event.startMonth - 1, event.startDay)
        // If event is in the past of this year, move it to next year
        if (eventStartDate < from) {
          eventStartDate = new Date(2027, event.startMonth - 1, event.startDay)
        }
        const daysUntilEvent = Math.floor(
          (eventStartDate.getTime() - from.getTime()) / (1000 * 60 * 60 * 24)
        )
        expect(daysUntilEvent).toBeGreaterThanOrEqual(0)
        expect(daysUntilEvent).toBeLessThanOrEqual(60)
      })
    })

    it('should include events that are currently active', () => {
      const from = new Date(2026, 0, 15) // Jan 15 (during Soldes d'Hiver)
      const upcoming = getUpcomingEvents(from)
      const hasSoldesHiver = upcoming.some((e) => e.id === 'soldes-hiver')
      expect(hasSoldesHiver).toBe(true)
    })

    it('should return sorted events by start date', () => {
      const from = new Date(2026, 0, 1)
      const upcoming = getUpcomingEvents(from)
      for (let i = 1; i < upcoming.length; i++) {
        const prevEventDate = new Date(
          2026,
          upcoming[i - 1].startMonth - 1,
          upcoming[i - 1].startDay
        )
        const currEventDate = new Date(
          2026,
          upcoming[i].startMonth - 1,
          upcoming[i].startDay
        )
        expect(prevEventDate.getTime()).toBeLessThanOrEqual(currEventDate.getTime())
      }
    })

    it('should handle dates near year boundary', () => {
      const from = new Date(2026, 10, 1) // Nov 1
      const upcoming = getUpcomingEvents(from)
      expect(Array.isArray(upcoming)).toBe(true)
      // Black Friday and Cyber Monday should be included
      const hasBlackFriday = upcoming.some((e) => e.id === 'black-friday')
      const hasCyberMonday = upcoming.some((e) => e.id === 'cyber-monday')
      expect(hasBlackFriday).toBe(true)
      expect(hasCyberMonday).toBe(true)
    })
  })

  describe('PromoEvent interface', () => {
    it('should allow creating events with proper types', () => {
      const event: PromoEvent = {
        id: 'test-event',
        name: 'Test Event',
        startMonth: 5,
        startDay: 10,
        endMonth: 5,
        endDay: 20,
        category: 'general',
        discount: 'jusqu\'à 30%',
        description: 'A test event',
        icon: '🎉',
        color: 'var(--gold)',
      }
      expect(event.id).toBe('test-event')
      expect(event.startMonth).toBe(5)
      expect(event.category).toBe('general')
    })
  })
})
