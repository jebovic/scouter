import { describe, it, expect } from 'vitest'
import { computePurchaseTimeline } from './purchaseTimeline'

describe('purchaseTimeline', () => {
  const referenceDate = new Date('2026-03-15') // March 15, 2026

  describe('priority determination', () => {
    it('returns priority="now" when currentPrice <= targetPrice', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 100,
          targetPrice: 100,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result).toHaveLength(1)
      expect(result[0].priority).toBe('now')
      expect(result[0].reason).toBe('Prix cible atteint')
    })

    it('returns priority="now" when currentPrice < targetPrice', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 95,
          targetPrice: 100,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('now')
    })

    it('returns priority="now" when dealScore >= 80', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 150,
          targetPrice: 100,
          dealScore: 80,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('now')
      expect(result[0].reason).toBe('Excellent deal score')
    })

    it('returns priority="soon" when dealScore >= 60 and < 80', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 150,
          targetPrice: 100,
          dealScore: 70,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('soon')
      expect(result[0].reason).toBe('Bon deal, à surveiller')
    })

    it('returns priority="wait" when dealScore < 60 and no target met', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 150,
          targetPrice: 100,
          dealScore: 50,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('wait')
      expect(result[0].reason).toBe('Attendre une meilleure opportunité')
    })

    it('returns priority="wait" when no dealScore and target not met', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 150,
          targetPrice: 100,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('wait')
    })

    it('returns priority="wait" when no targetPrice and dealScore < 60', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 100,
          dealScore: 50,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('wait')
    })

    it('returns priority="now" when no targetPrice but dealScore >= 80', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 100,
          dealScore: 85,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].priority).toBe('now')
    })
  })

  describe('window dates', () => {
    it('sets windowStart and windowEnd to today+7d for priority="now"', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 95,
          targetPrice: 100,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].windowStart.getTime()).toBe(referenceDate.getTime())
      // Check end date is approximately 7 days later
      const daysDiff = Math.floor((result[0].windowEnd.getTime() - referenceDate.getTime()) / (1000 * 60 * 60 * 24))
      expect(daysDiff).toBe(7)
    })

    it('sets windowStart=today+7d and windowEnd=today+30d for priority="soon"', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 150,
          targetPrice: 100,
          dealScore: 70,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      const startDiff = Math.round((result[0].windowStart.getTime() - referenceDate.getTime()) / (1000 * 60 * 60 * 24))
      const endDiff = Math.round((result[0].windowEnd.getTime() - referenceDate.getTime()) / (1000 * 60 * 60 * 24))
      expect(startDiff).toBe(7)
      expect(endDiff).toBe(30)
    })

    it('sets windowStart=today+30d and windowEnd=today+90d for priority="wait"', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 150,
          targetPrice: 100,
          dealScore: 50,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      const startDiff = Math.round((result[0].windowStart.getTime() - referenceDate.getTime()) / (1000 * 60 * 60 * 24))
      const endDiff = Math.round((result[0].windowEnd.getTime() - referenceDate.getTime()) / (1000 * 60 * 60 * 24))
      expect(startDiff).toBe(30)
      expect(endDiff).toBe(90)
    })
  })

  describe('data transformation', () => {
    it('returns empty array when items is empty', () => {
      const result = computePurchaseTimeline([], referenceDate)
      expect(result).toEqual([])
    })

    it('includes all required fields in TimelineItem', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 100,
          targetPrice: 100,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0]).toHaveProperty('itemId')
      expect(result[0]).toHaveProperty('itemName')
      expect(result[0]).toHaveProperty('merchant')
      expect(result[0]).toHaveProperty('currentPrice')
      expect(result[0]).toHaveProperty('targetPrice')
      expect(result[0]).toHaveProperty('windowStart')
      expect(result[0]).toHaveProperty('windowEnd')
      expect(result[0]).toHaveProperty('priority')
      expect(result[0]).toHaveProperty('reason')
    })

    it('processes multiple items correctly', () => {
      const items = [
        { id: '1', name: 'Item 1', merchant: 'Store A', price: 95, targetPrice: 100 },
        { id: '2', name: 'Item 2', merchant: 'Store B', price: 150, targetPrice: 100, dealScore: 70 },
        { id: '3', name: 'Item 3', merchant: 'Store C', price: 200, targetPrice: 100, dealScore: 30 },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result).toHaveLength(3)
      expect(result[0].priority).toBe('now')
      expect(result[1].priority).toBe('soon')
      expect(result[2].priority).toBe('wait')
    })

    it('handles items with null targetPrice', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 100,
          targetPrice: null,
          dealScore: 85,
        },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result).toHaveLength(1)
      expect(result[0].targetPrice).toBeUndefined()
    })

    it('uses current date when referenceDate is not provided', () => {
      const items = [
        {
          id: '1',
          name: 'Item 1',
          merchant: 'Store A',
          price: 95,
          targetPrice: 100,
        },
      ]
      const result = computePurchaseTimeline(items)
      const now = new Date()
      const daysDiff = Math.abs(
        Math.floor((result[0].windowStart.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
      )
      expect(daysDiff).toBeLessThan(1) // within same day
    })
  })

  describe('French i18n', () => {
    it('uses French reason strings', () => {
      const items = [
        { id: '1', name: 'Item 1', merchant: 'Store A', price: 95, targetPrice: 100 },
        { id: '2', name: 'Item 2', merchant: 'Store B', price: 150, dealScore: 80 },
        { id: '3', name: 'Item 3', merchant: 'Store C', price: 150, targetPrice: 100, dealScore: 70 },
        { id: '4', name: 'Item 4', merchant: 'Store D', price: 150, targetPrice: 100 },
      ]
      const result = computePurchaseTimeline(items, referenceDate)
      expect(result[0].reason).toBe('Prix cible atteint')
      expect(result[1].reason).toBe('Excellent deal score')
      expect(result[2].reason).toBe('Bon deal, à surveiller')
      expect(result[3].reason).toBe('Attendre une meilleure opportunité')
    })
  })
})
