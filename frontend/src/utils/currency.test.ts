import { describe, it, expect } from 'vitest'
import { formatCurrency, localeToLang } from './currency'

describe('formatCurrency', () => {
  it('formats EUR with fr-FR locale containing 1, 234, and €', () => {
    const result = formatCurrency(1234.5, 'EUR', 'fr-FR')
    expect(result).toContain('€')
    // French format: "1 234,50 €" or similar
    expect(result).toMatch(/1.*234/)
  })

  it('formats USD with en-US locale containing $', () => {
    const result = formatCurrency(1234.5, 'USD', 'en-US')
    expect(result).toContain('$')
    expect(result).toContain('1,234')
  })

  it('formats GBP with en-GB locale containing £', () => {
    const result = formatCurrency(1234.5, 'GBP', 'en-GB')
    expect(result).toContain('£')
  })

  it('formats zero correctly', () => {
    const result = formatCurrency(0, 'EUR', 'fr-FR')
    expect(result).toContain('0')
  })

  it('handles invalid currency by falling back to USD', () => {
    const result = formatCurrency(1234.5, 'INVALID', 'en-US')
    // Should fall back to USD format
    expect(result).toContain('$')
  })

  it('handles invalid locale by falling back to en-US', () => {
    const result = formatCurrency(1234.5, 'USD', 'invalid-XX')
    // Should still format as currency
    expect(result).toContain('$')
  })

  it('formats large numbers correctly', () => {
    const result = formatCurrency(99999.99, 'USD', 'en-US')
    expect(result).toBeTruthy()
    expect(typeof result).toBe('string')
  })

  it('formats negative numbers', () => {
    const result = formatCurrency(-1234.5, 'EUR', 'fr-FR')
    expect(result).toContain('€')
  })
})

describe('localeToLang', () => {
  it('maps fr-FR to fr', () => {
    expect(localeToLang('fr-FR')).toBe('fr')
  })

  it('maps en-US to en', () => {
    expect(localeToLang('en-US')).toBe('en')
  })

  it('maps en-GB to en', () => {
    expect(localeToLang('en-GB')).toBe('en')
  })

  it('maps de-DE to de', () => {
    expect(localeToLang('de-DE')).toBe('de')
  })

  it('maps unknown locale to en', () => {
    expect(localeToLang('unknown')).toBe('en')
  })

  it('maps any fr-prefixed locale to fr', () => {
    expect(localeToLang('fr-CA')).toBe('fr')
    expect(localeToLang('fr-BE')).toBe('fr')
  })

  it('maps any de-prefixed locale to de', () => {
    expect(localeToLang('de-AT')).toBe('de')
    expect(localeToLang('de-CH')).toBe('de')
  })

  it('defaults to en for any other locale', () => {
    expect(localeToLang('es-ES')).toBe('en')
    expect(localeToLang('it-IT')).toBe('en')
    expect(localeToLang('pt-BR')).toBe('en')
  })
})
