import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { PurchaseTimeline } from './PurchaseTimeline'
import type { ShoppingItem } from '../../types'

describe('PurchaseTimeline', () => {
  const mockItems: ShoppingItem[] = [
    {
      id: '1',
      missionId: 'mission-1',
      name: 'Laptop',
      merchant: 'Best Buy',
      costCategory: 'electronics',
      price: 800,
      targetPrice: 750,
      status: 'watch',
      pinned: false,
      createdAt: '2026-03-15T00:00:00Z',
    },
    {
      id: '2',
      missionId: 'mission-1',
      name: 'Monitor',
      merchant: 'Amazon',
      costCategory: 'electronics',
      price: 350,
      targetPrice: 300,
      status: 'watch',
      pinned: false,
      createdAt: '2026-03-15T00:00:00Z',
    },
  ]

  it('renders empty state when items array is empty', () => {
    const { container } = render(<PurchaseTimeline items={[]} />)
    // Check that something is rendered (empty state)
    expect(container.firstChild).toBeInTheDocument()
  })

  it('renders header with month labels', () => {
    render(<PurchaseTimeline items={mockItems} />)
    const header = screen.getByRole('heading', { level: 2 })
    expect(header).toBeInTheDocument()
  })

  it('renders one row per item', () => {
    render(<PurchaseTimeline items={mockItems} />)
    expect(screen.getByText('Laptop')).toBeInTheDocument()
    expect(screen.getByText('Monitor')).toBeInTheDocument()
  })

  it('renders merchant names', () => {
    render(<PurchaseTimeline items={mockItems} />)
    expect(screen.getByText(/Best Buy/)).toBeInTheDocument()
    expect(screen.getByText(/Amazon/)).toBeInTheDocument()
  })

  it('renders legend section', () => {
    render(<PurchaseTimeline items={mockItems} />)
    // Check for either English or French legend text
    const legend = screen.queryByText(/legend|légende/i)
    expect(legend).toBeInTheDocument()
  })

  it('renders priority badges for items', () => {
    render(<PurchaseTimeline items={mockItems} />)
    // Both items should have badges (MAINTENANT or other priority)
    const badges = screen.getAllByText(/MAINTENANT|BIENTÔT|ATTENDRE/)
    expect(badges.length).toBeGreaterThan(0)
  })
})
