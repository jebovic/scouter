import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import {
  convertCurrency,
  formatCurrency,
  fetchExchangeRates,
  SUPPORTED_CURRENCIES,
} from './currencyRates'

describe('currencyRates', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
  })

  afterEach(() => {
    localStorage.clear()
  })

  describe('convertCurrency', () => {
    const mockRates = {
      base: 'EUR' as const,
      rates: {
        EUR: 1.0,
        USD: 1.08,
        GBP: 0.86,
      },
      fetchedAt: Date.now(),
    }

    it('converts EUR to USD correctly', () => {
      const result = convertCurrency(100, 'EUR', 'USD', mockRates)
      expect(result).toBeCloseTo(108, 1)
    })

    it('converts USD to EUR correctly', () => {
      const result = convertCurrency(108, 'USD', 'EUR', mockRates)
      expect(result).toBeCloseTo(100, 1)
    })

    it('returns same amount when converting to same currency', () => {
      const result = convertCurrency(100, 'EUR', 'EUR', mockRates)
      expect(result).toBe(100)
    })

    it('handles zero amount', () => {
      const result = convertCurrency(0, 'EUR', 'USD', mockRates)
      expect(result).toBe(0)
    })

    it('handles currency pair via USD bridge when not directly available', () => {
      const result = convertCurrency(100, 'GBP', 'USD', mockRates)
      // GBP -> EUR: 100 / 0.86, EUR -> USD: * 1.08
      expect(result).toBeGreaterThan(100)
    })

    it('converts correctly with decimal amounts', () => {
      const result = convertCurrency(50.5, 'EUR', 'USD', mockRates)
      expect(result).toBeCloseTo(54.54, 1)
    })
  })

  describe('formatCurrency', () => {
    it('formats EUR with fr-FR locale', () => {
      const result = formatCurrency(1234.56, 'EUR', 'fr-FR')
      expect(result).toContain('€')
    })

    it('formats USD with en-US locale', () => {
      const result = formatCurrency(1234.56, 'USD', 'en-US')
      expect(result).toContain('$')
    })

    it('formats zero correctly', () => {
      const result = formatCurrency(0, 'EUR', 'fr-FR')
      expect(result).toContain('0')
    })

    it('formats large numbers', () => {
      const result = formatCurrency(99999.99, 'USD', 'en-US')
      expect(result).toBeTruthy()
      expect(result.length).toBeGreaterThan(0)
    })
  })

  describe('fetchExchangeRates', () => {
    it('returns cached rates if available and not stale', async () => {
      const mockRates = {
        base: 'EUR' as const,
        rates: { EUR: 1.0, USD: 1.08 },
        fetchedAt: Date.now(),
      }
      localStorage.setItem('scouter_fx_rates', JSON.stringify(mockRates))

      const result = await fetchExchangeRates()
      expect(result).toEqual(mockRates)
    })

    it('returns stale null for cached rates older than 4 hours', async () => {
      const oldRates = {
        base: 'EUR' as const,
        rates: { EUR: 1.0, USD: 1.08 },
        fetchedAt: Date.now() - 5 * 60 * 60 * 1000, // 5 hours ago
      }
      localStorage.setItem('scouter_fx_rates', JSON.stringify(oldRates))

      // Mock the fetch
      global.fetch = vi.fn(() =>
        Promise.resolve({
          ok: true,
          json: () =>
            Promise.resolve({
              data: {
                dataSets: [
                  {
                    series: {
                      '0:0': { observations: { '0': [1.1] } },
                      '1:0': { observations: { '0': [0.9] } },
                    },
                  },
                ],
              },
            }),
        } as Response)
      )

      const result = await fetchExchangeRates()
      expect(result.base).toBe('EUR')
    })
  })

  describe('SUPPORTED_CURRENCIES', () => {
    it('includes all expected currencies', () => {
      expect(SUPPORTED_CURRENCIES).toContain('EUR')
      expect(SUPPORTED_CURRENCIES).toContain('USD')
      expect(SUPPORTED_CURRENCIES).toContain('GBP')
      expect(SUPPORTED_CURRENCIES).toContain('CHF')
      expect(SUPPORTED_CURRENCIES).toContain('CAD')
      expect(SUPPORTED_CURRENCIES).toContain('JPY')
    })

    it('contains at least 10 currencies', () => {
      expect(SUPPORTED_CURRENCIES.length).toBeGreaterThanOrEqual(10)
    })
  })
})
