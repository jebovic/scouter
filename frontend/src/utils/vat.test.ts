import { describe, it, expect } from 'vitest'
import { htToTtc, ttcToHt, vatAmount, type VATRate } from './vat'

describe('vat utilities', () => {
  describe('htToTtc', () => {
    it('converts HT to TTC with 20% rate', () => {
      const result = htToTtc(100, 0.20)
      expect(result).toBe(120)
    })

    it('converts HT to TTC with 10% rate', () => {
      const result = htToTtc(100, 0.10)
      expect(result).toBeCloseTo(110, 5)
    })

    it('converts HT to TTC with 5.5% rate', () => {
      const result = htToTtc(100, 0.055)
      expect(result).toBe(105.5)
    })

    it('converts HT to TTC with 2.1% rate', () => {
      const result = htToTtc(100, 0.021)
      expect(result).toBeCloseTo(102.1, 5)
    })

    it('handles zero HT', () => {
      const result = htToTtc(0, 0.20)
      expect(result).toBe(0)
    })

    it('handles decimal HT values', () => {
      const result = htToTtc(99.99, 0.20)
      expect(result).toBeCloseTo(119.988, 5)
    })

    it('handles large HT values', () => {
      const result = htToTtc(10000, 0.20)
      expect(result).toBe(12000)
    })

    it('rounds to 2 decimals for display precision', () => {
      const result = htToTtc(19.99, 0.055)
      // 19.99 * 1.055 = 21.08945, should round to 21.09
      expect(result).toBeCloseTo(21.09, 2)
    })
  })

  describe('ttcToHt', () => {
    it('converts TTC back to HT with 20% rate', () => {
      const result = ttcToHt(120, 0.20)
      expect(result).toBeCloseTo(100, 5)
    })

    it('converts TTC back to HT with 10% rate', () => {
      const result = ttcToHt(110, 0.10)
      expect(result).toBeCloseTo(100, 5)
    })

    it('converts TTC back to HT with 5.5% rate', () => {
      const result = ttcToHt(105.5, 0.055)
      expect(result).toBeCloseTo(100, 5)
    })

    it('converts TTC back to HT with 2.1% rate', () => {
      const result = ttcToHt(102.1, 0.021)
      expect(result).toBeCloseTo(100, 5)
    })

    it('handles zero TTC', () => {
      const result = ttcToHt(0, 0.20)
      expect(result).toBe(0)
    })

    it('handles decimal TTC values', () => {
      const result = ttcToHt(119.988, 0.20)
      expect(result).toBeCloseTo(99.99, 5)
    })

    it('handles large TTC values', () => {
      const result = ttcToHt(12000, 0.20)
      expect(result).toBe(10000)
    })

    it('is the inverse of htToTtc', () => {
      const ht = 50
      const rate: VATRate = 0.20
      const ttc = htToTtc(ht, rate)
      const backToHt = ttcToHt(ttc, rate)
      expect(backToHt).toBeCloseTo(ht, 5)
    })
  })

  describe('vatAmount', () => {
    it('calculates VAT amount with 20% rate', () => {
      const result = vatAmount(100, 0.20)
      expect(result).toBe(20)
    })

    it('calculates VAT amount with 10% rate', () => {
      const result = vatAmount(100, 0.10)
      expect(result).toBe(10)
    })

    it('calculates VAT amount with 5.5% rate', () => {
      const result = vatAmount(100, 0.055)
      expect(result).toBe(5.5)
    })

    it('calculates VAT amount with 2.1% rate', () => {
      const result = vatAmount(100, 0.021)
      expect(result).toBeCloseTo(2.1, 5)
    })

    it('handles zero HT', () => {
      const result = vatAmount(0, 0.20)
      expect(result).toBe(0)
    })

    it('handles decimal HT values', () => {
      const result = vatAmount(99.99, 0.20)
      expect(result).toBeCloseTo(19.998, 5)
    })

    it('handles large HT values', () => {
      const result = vatAmount(10000, 0.20)
      expect(result).toBe(2000)
    })

    it('equals ttc - ht', () => {
      const ht = 75.50
      const rate: VATRate = 0.20
      const ttc = htToTtc(ht, rate)
      const vat = vatAmount(ht, rate)
      expect(vat).toBeCloseTo(ttc - ht, 5)
    })
  })

  describe('edge cases', () => {
    it('handles very small positive HT values', () => {
      const result = htToTtc(0.01, 0.20)
      expect(result).toBeCloseTo(0.012, 5)
    })

    it('preserves precision for currency calculations', () => {
      // Common use case: € 19.99 HT
      const ht = 19.99
      const ttc = htToTtc(ht, 0.20)
      const vat = vatAmount(ht, 0.20)
      expect(vat).toBeCloseTo(3.998, 5)
      expect(ttc).toBeCloseTo(ht + vat, 5)
    })

    it('handles all French VAT rates in sequence', () => {
      const ht = 100
      const rates: VATRate[] = [0.20, 0.10, 0.055, 0.021]
      const results = rates.map(r => htToTtc(ht, r))
      expect(results[0]).toBe(120)
      expect(results[1]).toBeCloseTo(110, 5)
      expect(results[2]).toBe(105.5)
      expect(results[3]).toBeCloseTo(102.1, 5)
    })
  })
})
