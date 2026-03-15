import { describe, it, expect } from 'vitest'
import { formatDate, formatCurrencyLocale, formatRelativeTime } from './format'

describe('formatDate', () => {
  it('formats ISO date string with en-US locale as MM/DD/YYYY style', () => {
    const result = formatDate('2024-01-15T00:00:00Z', 'en-US')
    expect(result).toMatch(/01.15.2024|01\/15\/2024/)
  })

  it('formats date with fr-FR locale in DD/MM/YYYY style', () => {
    const result = formatDate('2024-01-15T00:00:00Z', 'fr-FR')
    expect(result).toMatch(/15.01.2024|15\/01\/2024/)
  })

  it('accepts a Date object', () => {
    const date = new Date('2024-06-01T00:00:00Z')
    const result = formatDate(date, 'en-US')
    expect(result).toBeTruthy()
    expect(typeof result).toBe('string')
  })

  it('returns a non-empty string for valid inputs', () => {
    const result = formatDate('2024-12-31', 'en-US')
    expect(result.length).toBeGreaterThan(0)
  })
})

describe('formatCurrencyLocale', () => {
  it('formats EUR with fr-FR locale', () => {
    const result = formatCurrencyLocale(1234.56, 'fr-FR', 'EUR')
    // French format uses space or non-breaking space as thousands sep, € at end
    expect(result).toContain('€')
  })

  it('formats USD with en-US locale', () => {
    const result = formatCurrencyLocale(1234.56, 'en-US', 'USD')
    expect(result).toContain('$')
    expect(result).toContain('1,234')
  })

  it('formats zero correctly', () => {
    const result = formatCurrencyLocale(0, 'en-US', 'USD')
    expect(result).toContain('0')
  })

  it('formats large numbers', () => {
    const result = formatCurrencyLocale(99999.99, 'en-US', 'USD')
    expect(result).toBeTruthy()
  })

  it('formats GBP with en-GB locale', () => {
    const result = formatCurrencyLocale(500, 'en-GB', 'GBP')
    expect(result).toContain('£')
  })
})

describe('formatRelativeTime', () => {
  it('returns a non-empty string', () => {
    const future = new Date(Date.now() + 3600 * 1000).toISOString()
    const result = formatRelativeTime(future, 'en-US')
    expect(result.length).toBeGreaterThan(0)
  })

  it('handles past dates (hours ago)', () => {
    const past = new Date(Date.now() - 2 * 3600 * 1000).toISOString()
    const result = formatRelativeTime(past, 'en-US')
    expect(result).toMatch(/hour|hours/)
  })

  it('handles past dates (days ago)', () => {
    const past = new Date(Date.now() - 3 * 86400 * 1000).toISOString()
    const result = formatRelativeTime(past, 'en-US')
    expect(result).toMatch(/day|days/)
  })

  it('handles past dates (minutes ago)', () => {
    const past = new Date(Date.now() - 30 * 60 * 1000).toISOString()
    const result = formatRelativeTime(past, 'en-US')
    expect(result).toMatch(/minute|minutes/)
  })

  it('handles French locale', () => {
    const past = new Date(Date.now() - 2 * 3600 * 1000).toISOString()
    const result = formatRelativeTime(past, 'fr-FR')
    // French relative time should produce non-empty string
    expect(result.length).toBeGreaterThan(0)
  })

  it('accepts a Date object', () => {
    const date = new Date(Date.now() - 86400 * 1000)
    const result = formatRelativeTime(date, 'en-US')
    expect(result).toBeTruthy()
  })
})
