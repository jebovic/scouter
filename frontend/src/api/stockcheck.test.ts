import { describe, it, expect } from 'vitest'
import { StockStatusSchema } from './stockcheck'

describe('StockStatusSchema', () => {
  it('accepts in_stock', () => {
    const payload = { status: 'in_stock', retailer: 'fnac', updatedAt: new Date().toISOString() }
    expect(() => StockStatusSchema.parse(payload)).not.toThrow()
    expect(StockStatusSchema.parse(payload).status).toBe('in_stock')
  })

  it('accepts limited', () => {
    const payload = { status: 'limited', retailer: 'darty', updatedAt: new Date().toISOString() }
    expect(StockStatusSchema.parse(payload).status).toBe('limited')
  })

  it('accepts out_of_stock', () => {
    const payload = { status: 'out_of_stock', retailer: 'boulanger', updatedAt: new Date().toISOString() }
    expect(StockStatusSchema.parse(payload).status).toBe('out_of_stock')
  })

  it('rejects unknown status value', () => {
    const payload = { status: 'unknown', retailer: 'fnac', updatedAt: new Date().toISOString() }
    expect(() => StockStatusSchema.parse(payload)).toThrow()
  })

  it('rejects missing status', () => {
    const payload = { retailer: 'fnac', updatedAt: new Date().toISOString() }
    expect(() => StockStatusSchema.parse(payload)).toThrow()
  })

  it('rejects missing retailer', () => {
    const payload = { status: 'in_stock', updatedAt: new Date().toISOString() }
    expect(() => StockStatusSchema.parse(payload)).toThrow()
  })

  it('coerces updatedAt string to Date', () => {
    const payload = { status: 'in_stock', retailer: 'fnac', updatedAt: '2026-01-01T00:00:00Z' }
    const result = StockStatusSchema.parse(payload)
    expect(result.updatedAt).toBeInstanceOf(Date)
  })

  it('rejects missing updatedAt', () => {
    const payload = { status: 'in_stock', retailer: 'fnac' }
    expect(() => StockStatusSchema.parse(payload)).toThrow()
  })
})
