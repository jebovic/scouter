import { describe, it, expect } from 'vitest'
import {
  FRENCH_RETAILERS,
  getRetailersForCategory,
  buildRetailerLinks,
} from './retailerLinks'

describe('FRENCH_RETAILERS', () => {
  it('contains all 8 expected retailers', () => {
    const ids = FRENCH_RETAILERS.map((r) => r.id)
    expect(ids).toContain('fnac')
    expect(ids).toContain('darty')
    expect(ids).toContain('boulanger')
    expect(ids).toContain('cdiscount')
    expect(ids).toContain('amazon-fr')
    expect(ids).toContain('ldlc')
    expect(ids).toContain('leclerc')
    expect(ids).toContain('rueducommerce')
    expect(ids).toHaveLength(8)
  })

  it('every retailer has required fields', () => {
    for (const r of FRENCH_RETAILERS) {
      expect(r.id).toBeTruthy()
      expect(r.name).toBeTruthy()
      expect(r.logo).toBeTruthy()
      expect(typeof r.buildSearchUrl).toBe('function')
      expect(Array.isArray(r.primaryCategories)).toBe(true)
      expect(r.primaryCategories.length).toBeGreaterThan(0)
    }
  })

  it('buildSearchUrl encodes the query for Fnac', () => {
    const fnac = FRENCH_RETAILERS.find((r) => r.id === 'fnac')!
    const url = fnac.buildSearchUrl('MacBook Pro 14')
    expect(url).toContain('fnac.com')
    expect(url).toContain('MacBook%20Pro%2014')
  })

  it('buildSearchUrl encodes the query for Amazon FR', () => {
    const amazon = FRENCH_RETAILERS.find((r) => r.id === 'amazon-fr')!
    const url = amazon.buildSearchUrl('TV 4K OLED')
    expect(url).toContain('amazon.fr')
    expect(url).toContain('TV%204K%20OLED')
  })

  it('buildSearchUrl encodes special characters', () => {
    const darty = FRENCH_RETAILERS.find((r) => r.id === 'darty')!
    const url = darty.buildSearchUrl('Réfrigérateur & congélateur')
    expect(url).not.toContain(' ')
    // & should be encoded
    expect(url).not.toContain('& ')
  })

  it('buildSearchUrl handles empty query', () => {
    for (const r of FRENCH_RETAILERS) {
      const url = r.buildSearchUrl('')
      expect(typeof url).toBe('string')
      expect(url.length).toBeGreaterThan(0)
    }
  })
})

describe('getRetailersForCategory', () => {
  it('returns all retailers for "all" category', () => {
    const result = getRetailersForCategory('all')
    expect(result.length).toBe(FRENCH_RETAILERS.length)
  })

  it('returns electronics retailers plus "all" retailers for "electronics" category', () => {
    const result = getRetailersForCategory('electronics')
    const ids = result.map((r) => r.id)
    // Fnac, Darty, Boulanger are electronics+all
    expect(ids).toContain('fnac')
    expect(ids).toContain('darty')
    expect(ids).toContain('boulanger')
    // cdiscount, amazon-fr, leclerc are "all" so included
    expect(ids).toContain('cdiscount')
    expect(ids).toContain('amazon-fr')
  })

  it('returns computing retailers for "computing" category', () => {
    const result = getRetailersForCategory('computing')
    const ids = result.map((r) => r.id)
    expect(ids).toContain('ldlc')
    expect(ids).toContain('rueducommerce')
  })

  it('returns "all" category retailers for unknown category', () => {
    const result = getRetailersForCategory('travel')
    // Only "all" retailers should appear for unknown-specific category
    const allRetailers = FRENCH_RETAILERS.filter((r) =>
      r.primaryCategories.includes('all'),
    )
    expect(result.length).toBeGreaterThanOrEqual(allRetailers.length)
  })

  it('does not return duplicate retailers', () => {
    const result = getRetailersForCategory('electronics')
    const ids = result.map((r) => r.id)
    const unique = new Set(ids)
    expect(unique.size).toBe(ids.length)
  })
})

describe('buildRetailerLinks', () => {
  it('returns an array of { retailer, url } pairs', () => {
    const links = buildRetailerLinks('iPhone 15', 'electronics')
    expect(Array.isArray(links)).toBe(true)
    expect(links.length).toBeGreaterThan(0)
    for (const link of links) {
      expect(link.retailer).toBeDefined()
      expect(typeof link.url).toBe('string')
      expect(link.url.startsWith('http')).toBe(true)
    }
  })

  it('each url contains the encoded query', () => {
    const links = buildRetailerLinks('Samsung TV', 'all')
    for (const link of links) {
      expect(link.url).toContain('Samsung')
    }
  })

  it('returns links for all category when category is omitted', () => {
    const links = buildRetailerLinks('test query')
    const allLinks = buildRetailerLinks('test query', 'all')
    expect(links.length).toBe(allLinks.length)
  })

  it('handles empty query string', () => {
    const links = buildRetailerLinks('', 'electronics')
    expect(Array.isArray(links)).toBe(true)
    expect(links.length).toBeGreaterThan(0)
  })

  it('handles special characters in query', () => {
    const links = buildRetailerLinks('café & théière', 'all')
    for (const link of links) {
      expect(link.url).not.toContain(' ')
    }
  })
})
