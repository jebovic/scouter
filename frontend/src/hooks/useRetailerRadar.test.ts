import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { useRetailerRadar } from './useRetailerRadar'
import * as shoppingModule from '../api'
import { useQueryClient } from '@tanstack/react-query'
import type { ShoppingItem } from '../types'

// Mock the useShopping hook
vi.mock('./useShopping', () => ({
  useShopping: vi.fn(),
}))

import { useShopping } from './useShopping'

describe('useRetailerRadar', () => {
  it('groups items by merchant and picks lowest price', () => {
    const items: ShoppingItem[] = [
      {
        id: '1',
        missionId: 'mission-1',
        name: 'Laptop',
        merchant: 'Amazon',
        costCategory: 'electronics',
        price: 1200,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '2',
        missionId: 'mission-1',
        name: 'Monitor',
        merchant: 'Amazon',
        costCategory: 'electronics',
        price: 400,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '3',
        missionId: 'mission-1',
        name: 'Laptop',
        merchant: 'Decathlon',
        costCategory: 'electronics',
        price: 1100,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
    ]

    vi.mocked(useShopping).mockReturnValue({
      items,
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useRetailerRadar('mission-1'))

    expect(result.current.stats).toHaveLength(2)
    expect(result.current.stats[0].merchant).toBe('Amazon')
    expect(result.current.stats[0].bestPrice).toBe(400)
    expect(result.current.stats[1].merchant).toBe('Decathlon')
    expect(result.current.stats[1].bestPrice).toBe(1100)
  })

  it('sorts by price ascending', () => {
    const items: ShoppingItem[] = [
      {
        id: '1',
        missionId: 'mission-1',
        name: 'Item A',
        merchant: 'Store A',
        costCategory: 'other',
        price: 500,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '2',
        missionId: 'mission-1',
        name: 'Item B',
        merchant: 'Store B',
        costCategory: 'other',
        price: 200,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '3',
        missionId: 'mission-1',
        name: 'Item C',
        merchant: 'Store C',
        costCategory: 'other',
        price: 350,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
    ]

    vi.mocked(useShopping).mockReturnValue({
      items,
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useRetailerRadar('mission-1'))

    expect(result.current.stats.map((s) => s.bestPrice)).toEqual([200, 350, 500])
  })

  it('filters items with price <= 0', () => {
    const items: ShoppingItem[] = [
      {
        id: '1',
        missionId: 'mission-1',
        name: 'Item A',
        merchant: 'Store A',
        costCategory: 'other',
        price: 100,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '2',
        missionId: 'mission-1',
        name: 'Item B',
        merchant: 'Store B',
        costCategory: 'other',
        price: 0,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '3',
        missionId: 'mission-1',
        name: 'Item C',
        merchant: 'Store C',
        costCategory: 'other',
        price: -50,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
    ]

    vi.mocked(useShopping).mockReturnValue({
      items,
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useRetailerRadar('mission-1'))

    expect(result.current.stats).toHaveLength(1)
    expect(result.current.stats[0].merchant).toBe('Store A')
  })

  it('filters items without merchant', () => {
    const items: ShoppingItem[] = [
      {
        id: '1',
        missionId: 'mission-1',
        name: 'Item A',
        merchant: 'Store A',
        costCategory: 'other',
        price: 100,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
      {
        id: '2',
        missionId: 'mission-1',
        name: 'Item B',
        merchant: '',
        costCategory: 'other',
        price: 200,
        status: 'watch',
        pinned: false,
        createdAt: '2026-01-01',
      },
    ]

    vi.mocked(useShopping).mockReturnValue({
      items,
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useRetailerRadar('mission-1'))

    expect(result.current.stats).toHaveLength(1)
    expect(result.current.stats[0].merchant).toBe('Store A')
  })

  it('returns empty stats when no items', () => {
    vi.mocked(useShopping).mockReturnValue({
      items: [],
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useRetailerRadar('mission-1'))

    expect(result.current.stats).toHaveLength(0)
  })

  it('includes all item fields in retailer stats', () => {
    const items: ShoppingItem[] = [
      {
        id: 'item-123',
        missionId: 'mission-1',
        name: 'Gaming Laptop',
        merchant: 'Best Buy',
        costCategory: 'electronics',
        price: 1299,
        status: 'buy',
        pinned: false,
        createdAt: '2026-01-01',
      },
    ]

    vi.mocked(useShopping).mockReturnValue({
      items,
      isLoading: false,
      error: null,
    })

    const { result } = renderHook(() => useRetailerRadar('mission-1'))

    expect(result.current.stats[0]).toEqual({
      merchant: 'Best Buy',
      bestPrice: 1299,
      currency: 'EUR',
      itemName: 'Gaming Laptop',
      itemId: 'item-123',
      inStock: true,
    })
  })
})
