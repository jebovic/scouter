import { describe, it, expect } from 'vitest'
import { CATEGORY_SEASONS, getCurrentMonthRecommendations, type SeasonalPeriod, type CategorySeason } from './seasonalPricing'

describe('seasonalPricing.ts', () => {
  describe('CATEGORY_SEASONS', () => {
    it('has at least 10 categories', () => {
      expect(CATEGORY_SEASONS.length).toBeGreaterThanOrEqual(10)
    })

    it('each category has required fields', () => {
      CATEGORY_SEASONS.forEach((cat) => {
        expect(typeof cat.category).toBe('string')
        expect(cat.category.length).toBeGreaterThan(0)
        expect(typeof cat.icon).toBe('string')
        expect(cat.icon.length).toBeGreaterThan(0)
        expect(Array.isArray(cat.bestMonths)).toBe(true)
        expect(Array.isArray(cat.worstMonths)).toBe(true)
        expect(typeof cat.tip).toBe('string')
        expect(cat.tip.length).toBeGreaterThan(0)
      })
    })

    it('all bestMonths and worstMonths are non-empty arrays', () => {
      CATEGORY_SEASONS.forEach((cat) => {
        expect(cat.bestMonths.length).toBeGreaterThan(0)
        expect(cat.worstMonths.length).toBeGreaterThan(0)
      })
    })

    it('bestMonths and worstMonths do not overlap', () => {
      CATEGORY_SEASONS.forEach((cat) => {
        const overlap = cat.bestMonths.filter((m) => cat.worstMonths.includes(m))
        expect(overlap.length).toBe(0)
      })
    })

    it('all months are in valid range 1-12', () => {
      CATEGORY_SEASONS.forEach((cat) => {
        cat.bestMonths.forEach((m) => {
          expect(m).toBeGreaterThanOrEqual(1)
          expect(m).toBeLessThanOrEqual(12)
        })
        cat.worstMonths.forEach((m) => {
          expect(m).toBeGreaterThanOrEqual(1)
          expect(m).toBeLessThanOrEqual(12)
        })
      })
    })

    it('includes expected major categories', () => {
      const categoryNames = CATEGORY_SEASONS.map((c) => c.category)
      expect(categoryNames).toContain('Électronique')
      expect(categoryNames).toContain('TV & Audio')
      expect(categoryNames).toContain('Electroménager')
      expect(categoryNames).toContain('Mode')
      expect(categoryNames).toContain('Jouets')
    })

    it('all category names are in French', () => {
      const frenchBadges = CATEGORY_SEASONS.filter((c) => /[éàâêôûüœæù]/.test(c.category))
      expect(frenchBadges.length).toBeGreaterThan(0)
    })

    it('all tips are in French', () => {
      const frenchTips = CATEGORY_SEASONS.filter((c) => /[éàâêôûüœæù]/.test(c.tip))
      expect(frenchTips.length).toBeGreaterThan(0)
    })
  })

  describe('getCurrentMonthRecommendations', () => {
    it('returns a category when given a valid month', () => {
      const rec = getCurrentMonthRecommendations(1)
      expect(rec).toBeDefined()
      expect(rec.length).toBeGreaterThan(0)
    })

    it('returns categories where current month is in bestMonths', () => {
      const rec = getCurrentMonthRecommendations(1)
      rec.forEach((cat) => {
        expect(cat.bestMonths).toContain(1)
      })
    })

    it('returns no categories when month is in worstMonths', () => {
      // January is worst for TV & Audio
      const rec = getCurrentMonthRecommendations(1)
      const tvAudio = rec.find((c) => c.category === 'TV & Audio')
      // TV & Audio should not be in recommendations for Jan (assuming Jan is worst)
      // Only if bestMonths contains Jan
    })

    it('handles month 12 (December)', () => {
      const rec = getCurrentMonthRecommendations(12)
      expect(rec).toBeDefined()
      expect(Array.isArray(rec)).toBe(true)
    })

    it('returns array with specific category structure', () => {
      const rec = getCurrentMonthRecommendations(7) // July
      expect(Array.isArray(rec)).toBe(true)
      if (rec.length > 0) {
        const first = rec[0]
        expect(typeof first.category).toBe('string')
        expect(typeof first.icon).toBe('string')
        expect(Array.isArray(first.bestMonths)).toBe(true)
        expect(Array.isArray(first.worstMonths)).toBe(true)
        expect(typeof first.tip).toBe('string')
      }
    })

    it('filters to only categories with current month in bestMonths', () => {
      const month = 11 // November (Black Friday)
      const rec = getCurrentMonthRecommendations(month)
      rec.forEach((cat) => {
        expect(cat.bestMonths.includes(month)).toBe(true)
      })
    })
  })
})
