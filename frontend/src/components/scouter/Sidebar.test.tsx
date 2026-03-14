import { describe, it, expect, vi } from 'vitest'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { renderWithProviders } from '../../test/utils'
import { Sidebar } from './Sidebar'
import type { Mission } from '../../types'

const mockMissions: Mission[] = [
  {
    id: '1',
    slug: 'gaming-pc',
    name: 'Gaming PC',
    icon: '🖥️',
    category: 'electronics',
    phase: 'researching',
    budget: 2000,
    currency: 'USD',
    locale: 'en-US',
    constraints: [],
    costCategories: [],
    timeline: [],
    weightProfile: { price: 0.5, quality: 0.3, feature: 0.2 },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
  {
    id: '2',
    slug: 'sofa',
    name: 'New Sofa',
    icon: '🛋️',
    category: 'custom',
    phase: 'comparing',
    budget: 1000,
    currency: 'USD',
    locale: 'en-US',
    constraints: [],
    costCategories: [],
    timeline: [],
    weightProfile: { price: 0.5, quality: 0.3, feature: 0.2 },
    createdAt: '2024-01-02T00:00:00Z',
    updatedAt: '2024-01-02T00:00:00Z',
  },
]

describe('Sidebar', () => {
  it('renders nothing when collapsed', () => {
    renderWithProviders(
      <Sidebar open={false} missions={mockMissions} onClose={vi.fn()} />,
    )
    expect(screen.queryByRole('navigation', { name: /missions/i })).not.toBeInTheDocument()
  })

  it('renders mission list when open', () => {
    renderWithProviders(
      <Sidebar open missions={mockMissions} onClose={vi.fn()} />,
    )
    expect(screen.getByRole('navigation', { name: /missions/i })).toBeInTheDocument()
    expect(screen.getByText('Gaming PC')).toBeInTheDocument()
    expect(screen.getByText('New Sofa')).toBeInTheDocument()
  })

  it('calls onClose when backdrop is clicked', async () => {
    const onClose = vi.fn()
    renderWithProviders(
      <Sidebar open missions={mockMissions} onClose={onClose} />,
    )
    await userEvent.click(screen.getByRole('presentation'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('renders mission icons', () => {
    renderWithProviders(
      <Sidebar open missions={mockMissions} onClose={vi.fn()} />,
    )
    expect(screen.getByText('🖥️')).toBeInTheDocument()
    expect(screen.getByText('🛋️')).toBeInTheDocument()
  })

  it('renders empty state when no missions', () => {
    renderWithProviders(
      <Sidebar open missions={[]} onClose={vi.fn()} />,
    )
    expect(screen.getByText(/no missions/i)).toBeInTheDocument()
  })
})
