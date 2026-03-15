import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { RetailerRadar } from './RetailerRadar'

// Mock the useRetailerRadar hook
vi.mock('../../hooks/useRetailerRadar', () => ({
  useRetailerRadar: vi.fn(),
}))

import { useRetailerRadar } from '../../hooks/useRetailerRadar'
import type { RetailerStat } from '../../hooks/useRetailerRadar'

describe('RetailerRadar Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders empty state when no stats', () => {
    vi.mocked(useRetailerRadar).mockReturnValue({
      stats: [],
    })

    render(<RetailerRadar missionSlug="mission-123" />)

    expect(screen.getByText('Radar Revendeurs')).toBeInTheDocument()
    expect(screen.getByText('Aucun article trouvé')).toBeInTheDocument()
  })

  it('renders retailer stats with medals for top 3', () => {
    const stats: RetailerStat[] = [
      {
        merchant: 'Store A',
        bestPrice: 100,
        currency: 'EUR',
        itemName: 'Item A',
        itemId: 'item-1',
        inStock: true,
      },
      {
        merchant: 'Store B',
        bestPrice: 200,
        currency: 'EUR',
        itemName: 'Item B',
        itemId: 'item-2',
        inStock: true,
      },
      {
        merchant: 'Store C',
        bestPrice: 300,
        currency: 'EUR',
        itemName: 'Item C',
        itemId: 'item-3',
        inStock: true,
      },
    ]

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    render(<RetailerRadar missionSlug="mission-123" />)

    expect(screen.getByText('🥇')).toBeInTheDocument()
    expect(screen.getByText('🥈')).toBeInTheDocument()
    expect(screen.getByText('🥉')).toBeInTheDocument()
  })

  it('renders numbered ranks for items beyond top 3', () => {
    const stats: RetailerStat[] = Array.from({ length: 5 }, (_, i) => ({
      merchant: `Store ${String.fromCharCode(65 + i)}`,
      bestPrice: (i + 1) * 100,
      currency: 'EUR',
      itemName: `Item ${String.fromCharCode(65 + i)}`,
      itemId: `item-${i}`,
      inStock: true,
    }))

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    render(<RetailerRadar missionSlug="mission-123" />)

    expect(screen.getByText('4.')).toBeInTheDocument()
    expect(screen.getByText('5.')).toBeInTheDocument()
  })

  it('formats prices with currency symbol', () => {
    const stats: RetailerStat[] = [
      {
        merchant: 'Store A',
        bestPrice: 1299.5,
        currency: 'EUR',
        itemName: 'Item',
        itemId: 'item-1',
        inStock: true,
      },
    ]

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    const { container } = render(<RetailerRadar missionSlug="mission-123" />)

    // Check that price is rendered with currency symbol
    const priceElements = container.querySelectorAll('[class*="price"]')
    const hasPrice = Array.from(priceElements).some((el) => el.textContent?.includes('€'))
    expect(hasPrice).toBe(true)
  })

  it('highlights best price row', () => {
    const stats: RetailerStat[] = [
      {
        merchant: 'Store A',
        bestPrice: 100,
        currency: 'EUR',
        itemName: 'Item A',
        itemId: 'item-1',
        inStock: true,
      },
    ]

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    const { container } = render(<RetailerRadar missionSlug="mission-123" />)

    // Find the row with the medal (first row should have bestPrice class)
    const medalElement = screen.getByText('🥇')
    const rowElement = medalElement.closest('div[class*="row"]')
    expect(rowElement?.className).toContain('bestPrice')
  })

  it('shows out of stock badge when item not in stock', () => {
    const stats: RetailerStat[] = [
      {
        merchant: 'Store A',
        bestPrice: 100,
        currency: 'EUR',
        itemName: 'Item A',
        itemId: 'item-1',
        inStock: false,
      },
    ]

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    render(<RetailerRadar missionSlug="mission-123" />)

    expect(screen.getByText('Rupture')).toBeInTheDocument()
  })

  it('shows "Voir plus" button when more than 8 items', () => {
    const stats: RetailerStat[] = Array.from({ length: 10 }, (_, i) => ({
      merchant: `Store ${String.fromCharCode(65 + i)}`,
      bestPrice: (i + 1) * 100,
      currency: 'EUR',
      itemName: `Item ${String.fromCharCode(65 + i)}`,
      itemId: `item-${i}`,
      inStock: true,
    }))

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    render(<RetailerRadar missionSlug="mission-123" />)

    expect(screen.getByText(/Voir plus/)).toBeInTheDocument()
  })

  it('toggles show more state', async () => {
    const user = userEvent.setup()
    const stats: RetailerStat[] = Array.from({ length: 10 }, (_, i) => ({
      merchant: `Store ${String.fromCharCode(65 + i)}`,
      bestPrice: (i + 1) * 100,
      currency: 'EUR',
      itemName: `Item ${String.fromCharCode(65 + i)}`,
      itemId: `item-${i}`,
      inStock: true,
    }))

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    render(<RetailerRadar missionSlug="mission-123" />)

    const toggleButton = screen.getByRole('button', { name: /Voir plus/ })
    await user.click(toggleButton)

    expect(screen.getByRole('button', { name: /Afficher moins/ })).toBeInTheDocument()
  })

  it('truncates long item names', () => {
    const stats: RetailerStat[] = [
      {
        merchant: 'Store A',
        bestPrice: 100,
        currency: 'EUR',
        itemName: 'This is a very long item name that should be truncated',
        itemId: 'item-1',
        inStock: true,
      },
    ]

    vi.mocked(useRetailerRadar).mockReturnValue({ stats })

    render(<RetailerRadar missionSlug="mission-123" />)

    expect(screen.getByText(/This is a very long item na\.\.\./)).toBeInTheDocument()
  })
})
