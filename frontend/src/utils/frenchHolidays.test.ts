import { describe, it, expect } from 'vitest'
import {
  getFrenchHolidays,
  getUpcomingHolidays,
  isHoliday,
  type FrenchHoliday,
} from './frenchHolidays'

describe('frenchHolidays.ts', () => {
  describe('getFrenchHolidays', () => {
    it('returns array of holidays for a given year', () => {
      const holidays = getFrenchHolidays(2024)
      expect(Array.isArray(holidays)).toBe(true)
      expect(holidays.length).toBeGreaterThan(0)
    })

    it('contains fixed holiday: Jour de l\'An on 2024-01-01', () => {
      const holidays = getFrenchHolidays(2024)
      const jourAn = holidays.find((h) => h.date === '2024-01-01')
      expect(jourAn).toBeDefined()
      expect(jourAn?.name).toBe('Jour de l\'An')
      expect(jourAn?.nameEn).toBe('New Year\'s Day')
      expect(jourAn?.type).toBe('public')
    })

    it('contains fixed holiday: Fête du Travail on 2024-05-01 with stores closed', () => {
      const holidays = getFrenchHolidays(2024)
      const feteTravail = holidays.find((h) => h.date === '2024-05-01')
      expect(feteTravail).toBeDefined()
      expect(feteTravail?.name).toBe('Fête du Travail')
      expect(feteTravail?.storesLikelyClosed).toBe(true)
    })

    it('contains fixed holiday: Victoire 1945 on 2024-05-08', () => {
      const holidays = getFrenchHolidays(2024)
      const victoire = holidays.find((h) => h.date === '2024-05-08')
      expect(victoire).toBeDefined()
      expect(victoire?.name).toBe('Victoire 1945')
    })

    it('contains fixed holiday: Fête Nationale on 2024-07-14', () => {
      const holidays = getFrenchHolidays(2024)
      const feteNat = holidays.find((h) => h.date === '2024-07-14')
      expect(feteNat).toBeDefined()
      expect(feteNat?.name).toBe('Fête Nationale')
    })

    it('contains fixed holiday: Assomption on 2024-08-15', () => {
      const holidays = getFrenchHolidays(2024)
      const assomption = holidays.find((h) => h.date === '2024-08-15')
      expect(assomption).toBeDefined()
      expect(assomption?.name).toBe('Assomption')
    })

    it('contains fixed holiday: Toussaint on 2024-11-01 with stores likely closed', () => {
      const holidays = getFrenchHolidays(2024)
      const toussaint = holidays.find((h) => h.date === '2024-11-01')
      expect(toussaint).toBeDefined()
      expect(toussaint?.name).toBe('Toussaint')
      expect(toussaint?.storesLikelyClosed).toBe(true)
    })

    it('contains fixed holiday: Armistice on 2024-11-11', () => {
      const holidays = getFrenchHolidays(2024)
      const armistice = holidays.find((h) => h.date === '2024-11-11')
      expect(armistice).toBeDefined()
      expect(armistice?.name).toBe('Armistice')
    })

    it('contains fixed holiday: Noël on 2024-12-25 with stores closed', () => {
      const holidays = getFrenchHolidays(2024)
      const noel = holidays.find((h) => h.date === '2024-12-25')
      expect(noel).toBeDefined()
      expect(noel?.name).toBe('Noël')
      expect(noel?.storesLikelyClosed).toBe(true)
    })

    it('contains Easter Monday 2024 on 2024-04-01', () => {
      const holidays = getFrenchHolidays(2024)
      const easterMonday = holidays.find((h) => h.name === 'Lundi de Pâques')
      expect(easterMonday).toBeDefined()
      expect(easterMonday?.date).toBe('2024-04-01')
    })

    it('contains Ascension 2024 on 2024-05-09', () => {
      const holidays = getFrenchHolidays(2024)
      const ascension = holidays.find((h) => h.name === 'Ascension')
      expect(ascension).toBeDefined()
      expect(ascension?.date).toBe('2024-05-09')
    })

    it('contains Pentecost Monday 2024 on 2024-05-20', () => {
      const holidays = getFrenchHolidays(2024)
      const pentecostMonday = holidays.find((h) => h.name === 'Lundi de Pentecôte')
      expect(pentecostMonday).toBeDefined()
      expect(pentecostMonday?.date).toBe('2024-05-20')
    })

    it('returns holidays sorted by date', () => {
      const holidays = getFrenchHolidays(2024)
      for (let i = 0; i < holidays.length - 1; i++) {
        expect(holidays[i].date <= holidays[i + 1].date).toBe(true)
      }
    })

    it('all holidays have required fields', () => {
      const holidays = getFrenchHolidays(2024)
      holidays.forEach((h) => {
        expect(h.date).toBeDefined()
        expect(/^\d{4}-\d{2}-\d{2}$/.test(h.date)).toBe(true)
        expect(h.name).toBeDefined()
        expect(h.nameEn).toBeDefined()
        expect(['public', 'school_break']).toContain(h.type)
        expect(typeof h.storesLikelyClosed).toBe('boolean')
      })
    })

    it('works for different years (2025)', () => {
      const holidays2025 = getFrenchHolidays(2025)
      expect(Array.isArray(holidays2025)).toBe(true)
      expect(holidays2025.length).toBeGreaterThan(0)
      // Verify at least Jour de l'An is present
      const jourAn = holidays2025.find((h) => h.date === '2025-01-01')
      expect(jourAn).toBeDefined()
    })
  })

  describe('getUpcomingHolidays', () => {
    it('returns array with default max 5 items', () => {
      const upcoming = getUpcomingHolidays()
      expect(Array.isArray(upcoming)).toBe(true)
      expect(upcoming.length).toBeLessThanOrEqual(5)
    })

    it('respects custom count parameter', () => {
      const upcoming = getUpcomingHolidays(3)
      expect(upcoming.length).toBeLessThanOrEqual(3)
    })

    it('returns only holidays in the future (>= today)', () => {
      const today = new Date()
      const todayStr = today.toISOString().slice(0, 10)
      const upcoming = getUpcomingHolidays(10)
      upcoming.forEach((h) => {
        expect(h.date >= todayStr).toBe(true)
      })
    })

    it('returns holidays sorted by date (earliest first)', () => {
      const upcoming = getUpcomingHolidays(10)
      for (let i = 0; i < upcoming.length - 1; i++) {
        expect(upcoming[i].date <= upcoming[i + 1].date).toBe(true)
      }
    })

    it('can return 0 items if max is 0', () => {
      const upcoming = getUpcomingHolidays(0)
      expect(upcoming.length).toBe(0)
    })

    it('includes holidays from next year if needed', () => {
      // This test ensures the function looks ahead into the next year
      const upcoming = getUpcomingHolidays(20)
      const today = new Date()
      const daysUntilNewYear = new Date(today.getFullYear() + 1, 0, 1).getTime() - today.getTime()
      const monthsUntilNewYear = daysUntilNewYear / (1000 * 60 * 60 * 24 * 30)

      // If we're in the last 6 months of the year and asking for many holidays, we should get next year's holidays
      if (monthsUntilNewYear < 6 && upcoming.length > 0) {
        const hasNextYearHoliday = upcoming.some((h) => h.date.startsWith((today.getFullYear() + 1).toString()))
        expect(hasNextYearHoliday).toBe(true)
      }
    })
  })

  describe('isHoliday', () => {
    it('returns holiday object for a public holiday', () => {
      const date = new Date(2024, 0, 1) // January 1, 2024
      const holiday = isHoliday(date)
      expect(holiday).toBeDefined()
      expect(holiday?.name).toBe('Jour de l\'An')
    })

    it('returns undefined for a non-holiday date', () => {
      const date = new Date(2024, 2, 15) // March 15, 2024 (random Tuesday)
      const holiday = isHoliday(date)
      expect(holiday).toBeUndefined()
    })

    it('finds Easter Monday 2024', () => {
      const date = new Date(2024, 3, 1) // April 1, 2024 (Easter Monday)
      const holiday = isHoliday(date)
      expect(holiday).toBeDefined()
      expect(holiday?.name).toBe('Lundi de Pâques')
    })

    it('finds Ascension 2024', () => {
      const date = new Date(2024, 4, 9) // May 9, 2024 (Ascension)
      const holiday = isHoliday(date)
      expect(holiday).toBeDefined()
      expect(holiday?.name).toBe('Ascension')
    })

    it('returns undefined when no holiday matches', () => {
      const date = new Date(2024, 11, 24) // December 24, 2024
      const holiday = isHoliday(date)
      expect(holiday).toBeUndefined()
    })

    it('works across different years', () => {
      const date2025 = new Date(2025, 0, 1) // January 1, 2025
      const holiday = isHoliday(date2025)
      expect(holiday).toBeDefined()
      expect(holiday?.name).toBe('Jour de l\'An')
    })
  })
})
